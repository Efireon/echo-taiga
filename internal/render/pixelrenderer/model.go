package pixelrenderer

import (
	"encoding/json"
	"errors"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/yourusername/echo-taiga/internal/engine/ecs"
)

// Vertex представляет вершину 3D-модели
type Vertex struct {
	X, Y, Z float64
}

// TexCoord представляет текстурные координаты
type TexCoord struct {
	X, Y float64
}

// VertexData содержит данные преобразованной вершины для рендеринга
type VertexData struct {
	ScreenX, ScreenY float64 // Координаты экрана
	Depth            float64 // Глубина для Z-буфера
	W                float64 // W-компонент для перспективного деления
}

// Triangle представляет треугольник в 3D-модели
type Triangle struct {
	V1, V2, V3    Vertex      // Вершины
	UV1, UV2, UV3 TexCoord    // Текстурные координаты
	Normal        ecs.Vector3 // Нормаль к поверхности
}

// Model представляет 3D-модель
type Model struct {
	Triangles []Triangle    // Треугольники модели
	Vertices  []Vertex      // Все вершины модели
	Normals   []ecs.Vector3 // Нормали вершин
	TexCoords []TexCoord    // Текстурные координаты
	Indices   []uint32      // Индексы вершин для треугольников

	// Метаданные модели
	Name        string
	Description string

	// Bounding box для быстрой отбраковки
	BoundingBox struct {
		Min, Max Vertex
	}
}

// ModelLoader отвечает за загрузку 3D-моделей
type ModelLoader interface {
	LoadModel(filename string) (*Model, error)
}

// SimpleModelLoader реализует простой загрузчик моделей
type SimpleModelLoader struct {
	BasePath string
}

// NewSimpleModelLoader создает новый загрузчик моделей
func NewSimpleModelLoader(basePath string) *SimpleModelLoader {
	return &SimpleModelLoader{
		BasePath: basePath,
	}
}

// LoadModel загружает модель из файла
func (ml *SimpleModelLoader) LoadModel(filename string) (*Model, error) {
	// Формируем полный путь к файлу
	fullPath := filepath.Join(ml.BasePath, filename)

	// Определяем формат файла по расширению
	ext := filepath.Ext(fullPath)

	switch ext {
	case ".obj":
		return ml.loadOBJ(fullPath)
	case ".json":
		return ml.loadJSON(fullPath)
	default:
		return nil, errors.New("unsupported model format: " + ext)
	}
}

// loadOBJ загружает модель в формате OBJ
func (ml *SimpleModelLoader) loadOBJ(filename string) (*Model, error) {
	// Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем содержимое файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Парсим OBJ формат
	model := &Model{
		Name: filepath.Base(filename),
	}

	// Временные слайсы для данных
	var vertices []Vertex
	var normals []ecs.Vector3
	var texCoords []TexCoord

	// Парсим строки файла
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue // Пропускаем пустые строки и комментарии
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "v": // Вершина
			if len(fields) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(fields[1], 64)
			y, _ := strconv.ParseFloat(fields[2], 64)
			z, _ := strconv.ParseFloat(fields[3], 64)
			vertices = append(vertices, Vertex{X: x, Y: y, Z: z})

			// Обновляем bounding box
			if len(vertices) == 1 {
				model.BoundingBox.Min = vertices[0]
				model.BoundingBox.Max = vertices[0]
			} else {
				// Обновляем минимальные значения
				if x < model.BoundingBox.Min.X {
					model.BoundingBox.Min.X = x
				}
				if y < model.BoundingBox.Min.Y {
					model.BoundingBox.Min.Y = y
				}
				if z < model.BoundingBox.Min.Z {
					model.BoundingBox.Min.Z = z
				}

				// Обновляем максимальные значения
				if x > model.BoundingBox.Max.X {
					model.BoundingBox.Max.X = x
				}
				if y > model.BoundingBox.Max.Y {
					model.BoundingBox.Max.Y = y
				}
				if z > model.BoundingBox.Max.Z {
					model.BoundingBox.Max.Z = z
				}
			}

		case "vn": // Нормаль
			if len(fields) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(fields[1], 64)
			y, _ := strconv.ParseFloat(fields[2], 64)
			z, _ := strconv.ParseFloat(fields[3], 64)
			normals = append(normals, ecs.Vector3{X: x, Y: y, Z: z})

		case "vt": // Текстурная координата
			if len(fields) < 3 {
				continue
			}
			u, _ := strconv.ParseFloat(fields[1], 64)
			v, _ := strconv.ParseFloat(fields[2], 64)
			texCoords = append(texCoords, TexCoord{X: u, Y: v})

		case "f": // Грань (треугольник)
			if len(fields) < 4 {
				continue
			}

			// Получаем индексы для треугольника
			var vertIndices [3]int
			var texIndices [3]int
			var normIndices [3]int

			for i := 0; i < 3; i++ {
				// Парсим индексы в формате v/vt/vn
				indices := strings.Split(fields[i+1], "/")

				if len(indices) >= 1 {
					idx, _ := strconv.Atoi(indices[0])
					vertIndices[i] = idx - 1 // Индексы в OBJ начинаются с 1
				}

				if len(indices) >= 2 && indices[1] != "" {
					idx, _ := strconv.Atoi(indices[1])
					texIndices[i] = idx - 1
				}

				if len(indices) >= 3 {
					idx, _ := strconv.Atoi(indices[2])
					normIndices[i] = idx - 1
				}
			}

			// Создаем треугольник
			var triangle Triangle

			// Устанавливаем вершины
			if vertIndices[0] >= 0 && vertIndices[0] < len(vertices) {
				triangle.V1 = vertices[vertIndices[0]]
			}
			if vertIndices[1] >= 0 && vertIndices[1] < len(vertices) {
				triangle.V2 = vertices[vertIndices[1]]
			}
			if vertIndices[2] >= 0 && vertIndices[2] < len(vertices) {
				triangle.V3 = vertices[vertIndices[2]]
			}

			// Устанавливаем текстурные координаты
			if texIndices[0] >= 0 && texIndices[0] < len(texCoords) {
				triangle.UV1 = texCoords[texIndices[0]]
			}
			if texIndices[1] >= 0 && texIndices[1] < len(texCoords) {
				triangle.UV2 = texCoords[texIndices[1]]
			}
			if texIndices[2] >= 0 && texIndices[2] < len(texCoords) {
				triangle.UV3 = texCoords[texIndices[2]]
			}

			// Устанавливаем нормаль
			if normIndices[0] >= 0 && normIndices[0] < len(normals) {
				triangle.Normal = normals[normIndices[0]]
			} else {
				// Если нормаль не указана, вычисляем её
				edge1 := ecs.Vector3{
					X: triangle.V2.X - triangle.V1.X,
					Y: triangle.V2.Y - triangle.V1.Y,
					Z: triangle.V2.Z - triangle.V1.Z,
				}

				edge2 := ecs.Vector3{
					X: triangle.V3.X - triangle.V1.X,
					Y: triangle.V3.Y - triangle.V1.Y,
					Z: triangle.V3.Z - triangle.V1.Z,
				}

				triangle.Normal = edge1.Cross(edge2).Normalize()
			}

			// Добавляем треугольник в модель
			model.Triangles = append(model.Triangles, triangle)
		}
	}

	model.Vertices = vertices
	model.Normals = normals
	model.TexCoords = texCoords

	return model, nil
}

// loadJSON загружает модель в формате JSON
func (ml *SimpleModelLoader) loadJSON(filename string) (*Model, error) {
	// Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Читаем содержимое файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Парсим JSON
	model := &Model{}
	err = json.Unmarshal(data, model)
	if err != nil {
		return nil, err
	}

	// Проверяем, есть ли у всех треугольников нормали
	for i := range model.Triangles {
		triangle := &model.Triangles[i]

		// Если нормаль не указана, вычисляем её
		if triangle.Normal.X == 0 && triangle.Normal.Y == 0 && triangle.Normal.Z == 0 {
			edge1 := ecs.Vector3{
				X: triangle.V2.X - triangle.V1.X,
				Y: triangle.V2.Y - triangle.V1.Y,
				Z: triangle.V2.Z - triangle.V1.Z,
			}

			edge2 := ecs.Vector3{
				X: triangle.V3.X - triangle.V1.X,
				Y: triangle.V3.Y - triangle.V1.Y,
				Z: triangle.V3.Z - triangle.V1.Z,
			}

			triangle.Normal = edge1.Cross(edge2).Normalize()
		}
	}

	// Если bounding box не указан, вычисляем его
	if model.BoundingBox.Min.X == 0 && model.BoundingBox.Min.Y == 0 && model.BoundingBox.Min.Z == 0 &&
		model.BoundingBox.Max.X == 0 && model.BoundingBox.Max.Y == 0 && model.BoundingBox.Max.Z == 0 {

		// Находим все вершины
		var vertices []Vertex
		for _, triangle := range model.Triangles {
			vertices = append(vertices, triangle.V1, triangle.V2, triangle.V3)
		}

		// Вычисляем bounding box
		if len(vertices) > 0 {
			model.BoundingBox.Min = vertices[0]
			model.BoundingBox.Max = vertices[0]

			for _, vertex := range vertices {
				// Обновляем минимальные значения
				if vertex.X < model.BoundingBox.Min.X {
					model.BoundingBox.Min.X = vertex.X
				}
				if vertex.Y < model.BoundingBox.Min.Y {
					model.BoundingBox.Min.Y = vertex.Y
				}
				if vertex.Z < model.BoundingBox.Min.Z {
					model.BoundingBox.Min.Z = vertex.Z
				}

				// Обновляем максимальные значения
				if vertex.X > model.BoundingBox.Max.X {
					model.BoundingBox.Max.X = vertex.X
				}
				if vertex.Y > model.BoundingBox.Max.Y {
					model.BoundingBox.Max.Y = vertex.Y
				}
				if vertex.Z > model.BoundingBox.Max.Z {
					model.BoundingBox.Max.Z = vertex.Z
				}
			}
		}
	}

	return model, nil
}

// TextureLoader отвечает за загрузку текстур
type TextureLoader interface {
	LoadTexture(filename string) (*ebiten.Image, error)
}

// SimpleTextureLoader реализует простой загрузчик текстур
type SimpleTextureLoader struct {
	BasePath string
}

// NewSimpleTextureLoader создает новый загрузчик текстур
func NewSimpleTextureLoader(basePath string) *SimpleTextureLoader {
	return &SimpleTextureLoader{
		BasePath: basePath,
	}
}

// LoadTexture загружает текстуру из файла
func (tl *SimpleTextureLoader) LoadTexture(filename string) (*ebiten.Image, error) {
	// Формируем полный путь к файлу
	fullPath := filepath.Join(tl.BasePath, filename)

	// Загружаем изображение
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Декодируем изображение
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Создаем текстуру Ebiten
	texture := ebiten.NewImageFromImage(img)

	return texture, nil
}

// Создание базовых примитивов

// createCubeModel создает модель куба
func createCubeModel() *Model {
	model := &Model{
		Name:        "cube",
		Description: "Basic cube model",
	}

	// Вершины куба
	vertices := []Vertex{
		{-0.5, -0.5, -0.5}, // 0: заднийнижнийлевый
		{0.5, -0.5, -0.5},  // 1: заднийнижнийправый
		{0.5, 0.5, -0.5},   // 2: заднийверхнийправый
		{-0.5, 0.5, -0.5},  // 3: заднийверхнийлевый
		{-0.5, -0.5, 0.5},  // 4: переднийнижнийлевый
		{0.5, -0.5, 0.5},   // 5: переднийнижнийправый
		{0.5, 0.5, 0.5},    // 6: переднийверхнийправый
		{-0.5, 0.5, 0.5},   // 7: переднийверхнийлевый
	}

	// Текстурные координаты
	texCoords := []TexCoord{
		{0, 0}, // 0: нижнийлевый
		{1, 0}, // 1: нижнийправый
		{1, 1}, // 2: верхнийправый
		{0, 1}, // 3: верхнийлевый
	}

	// Нормали для граней куба
	normals := []ecs.Vector3{
		{0, 0, -1}, // Задняя грань
		{0, 0, 1},  // Передняя грань
		{1, 0, 0},  // Правая грань
		{-1, 0, 0}, // Левая грань
		{0, 1, 0},  // Верхняя грань
		{0, -1, 0}, // Нижняя грань
	}

	// Создаем треугольники для каждой грани куба

	// Задняя грань (z = -0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[1], V3: vertices[2],
		UV1: texCoords[0], UV2: texCoords[1], UV3: texCoords[2],
		Normal: normals[0],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[2], V3: vertices[3],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[3],
		Normal: normals[0],
	})

	// Передняя грань (z = 0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[4], V2: vertices[6], V3: vertices[5],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[1],
		Normal: normals[1],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[4], V2: vertices[7], V3: vertices[6],
		UV1: texCoords[0], UV2: texCoords[3], UV3: texCoords[2],
		Normal: normals[1],
	})

	// Правая грань (x = 0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[1], V2: vertices[5], V3: vertices[6],
		UV1: texCoords[0], UV2: texCoords[1], UV3: texCoords[2],
		Normal: normals[2],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[1], V2: vertices[6], V3: vertices[2],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[3],
		Normal: normals[2],
	})

	// Левая грань (x = -0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[3], V3: vertices[7],
		UV1: texCoords[0], UV2: texCoords[3], UV3: texCoords[2],
		Normal: normals[3],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[7], V3: vertices[4],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[1],
		Normal: normals[3],
	})

	// Верхняя грань (y = 0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[3], V2: vertices[2], V3: vertices[6],
		UV1: texCoords[0], UV2: texCoords[1], UV3: texCoords[2],
		Normal: normals[4],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[3], V2: vertices[6], V3: vertices[7],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[3],
		Normal: normals[4],
	})

	// Нижняя грань (y = -0.5)
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[4], V3: vertices[5],
		UV1: texCoords[0], UV2: texCoords[1], UV3: texCoords[2],
		Normal: normals[5],
	})
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[5], V3: vertices[1],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[3],
		Normal: normals[5],
	})

	// Устанавливаем bounding box
	model.BoundingBox.Min = Vertex{-0.5, -0.5, -0.5}
	model.BoundingBox.Max = Vertex{0.5, 0.5, 0.5}

	return model
}

// createSphereModel создает модель сферы
func createSphereModel(segments, rings int) *Model {
	model := &Model{
		Name:        "sphere",
		Description: "Sphere model",
	}

	// Массивы для хранения вершин и текстурных координат
	var vertices []Vertex
	var texCoords []TexCoord

	// Создаем вершины сферы
	for i := 0; i <= rings; i++ {
		v := float64(i) / float64(rings)
		phi := v * math.Pi

		for j := 0; j <= segments; j++ {
			u := float64(j) / float64(segments)
			theta := u * 2 * math.Pi

			// Вычисляем координаты вершины
			x := math.Sin(phi) * math.Cos(theta)
			y := math.Cos(phi)
			z := math.Sin(phi) * math.Sin(theta)

			// Добавляем вершину
			vertices = append(vertices, Vertex{X: x * 0.5, Y: y * 0.5, Z: z * 0.5})

			// Добавляем текстурную координату
			texCoords = append(texCoords, TexCoord{X: u, Y: v})
		}
	}

	// Создаем треугольники
	for i := 0; i < rings; i++ {
		for j := 0; j < segments; j++ {
			// Индексы вершин текущего квадрата
			idx1 := i*(segments+1) + j
			idx2 := idx1 + 1
			idx3 := (i+1)*(segments+1) + j
			idx4 := idx3 + 1

			// Первый треугольник
			v1 := vertices[idx1]
			v2 := vertices[idx2]
			v3 := vertices[idx3]
			uv1 := texCoords[idx1]
			uv2 := texCoords[idx2]
			uv3 := texCoords[idx3]

			// Вычисляем нормаль к треугольнику
			normal1 := ecs.Vector3{
				X: v1.X * 2,
				Y: v1.Y * 2,
				Z: v1.Z * 2,
			}.Normalize()

			model.Triangles = append(model.Triangles, Triangle{
				V1: v1, V2: v2, V3: v3,
				UV1: uv1, UV2: uv2, UV3: uv3,
				Normal: normal1,
			})

			// Второй треугольник
			v1 = vertices[idx2]
			v2 = vertices[idx4]
			v3 = vertices[idx3]
			uv1 = texCoords[idx2]
			uv2 = texCoords[idx4]
			uv3 = texCoords[idx3]

			// Вычисляем нормаль к треугольнику
			normal2 := ecs.Vector3{
				X: v1.X * 2,
				Y: v1.Y * 2,
				Z: v1.Z * 2,
			}.Normalize()

			model.Triangles = append(model.Triangles, Triangle{
				V1: v1, V2: v2, V3: v3,
				UV1: uv1, UV2: uv2, UV3: uv3,
				Normal: normal2,
			})
		}
	}

	// Устанавливаем bounding box
	model.BoundingBox.Min = Vertex{-0.5, -0.5, -0.5}
	model.BoundingBox.Max = Vertex{0.5, 0.5, 0.5}

	return model
}

// createPlaneModel создает модель плоскости
func createPlaneModel() *Model {
	model := &Model{
		Name:        "plane",
		Description: "Plane model",
	}

	// Вершины плоскости
	vertices := []Vertex{
		{-0.5, 0, -0.5}, // 0: задний левый
		{0.5, 0, -0.5},  // 1: задний правый
		{0.5, 0, 0.5},   // 2: передний правый
		{-0.5, 0, 0.5},  // 3: передний левый
	}

	// Текстурные координаты
	texCoords := []TexCoord{
		{0, 0}, // 0: нижний левый
		{1, 0}, // 1: нижний правый
		{1, 1}, // 2: верхний правый
		{0, 1}, // 3: верхний левый
	}

	// Нормаль для плоскости (направлена вверх)
	normal := ecs.Vector3{0, 1, 0}

	// Создаем треугольники для плоскости
	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[1], V3: vertices[2],
		UV1: texCoords[0], UV2: texCoords[1], UV3: texCoords[2],
		Normal: normal,
	})

	model.Triangles = append(model.Triangles, Triangle{
		V1: vertices[0], V2: vertices[2], V3: vertices[3],
		UV1: texCoords[0], UV2: texCoords[2], UV3: texCoords[3],
		Normal: normal,
	})

	// Устанавливаем bounding box
	model.BoundingBox.Min = Vertex{-0.5, 0, -0.5}
	model.BoundingBox.Max = Vertex{0.5, 0, 0.5}

	return model
}

// createCylinderModel создает модель цилиндра
func createCylinderModel(segments int) *Model {
	model := &Model{
		Name:        "cylinder",
		Description: "Cylinder model",
	}

	// Массивы для хранения вершин
	var vertices []Vertex
	var topVertices []Vertex
	var bottomVertices []Vertex
	var texCoords []TexCoord

	// Центр верхнего основания
	topCenter := Vertex{0, 0.5, 0}

	// Центр нижнего основания
	bottomCenter := Vertex{0, -0.5, 0}

	// Создаем вершины боковой поверхности цилиндра
	for i := 0; i <= segments; i++ {
		angle := float64(i) / float64(segments) * 2 * math.Pi
		x := math.Cos(angle) * 0.5
		z := math.Sin(angle) * 0.5

		// Вершина на верхнем основании
		topVertices = append(topVertices, Vertex{X: x, Y: 0.5, Z: z})

		// Вершина на нижнем основании
		bottomVertices = append(bottomVertices, Vertex{X: x, Y: -0.5, Z: z})

		// Текстурная координата
		u := float64(i) / float64(segments)
		texCoords = append(texCoords, TexCoord{X: u, Y: 0})
		texCoords = append(texCoords, TexCoord{X: u, Y: 1})
	}

	// Создаем треугольники для боковой поверхности цилиндра
	for i := 0; i < segments; i++ {
		// Индексы вершин
		idx1 := i
		idx2 := (i + 1) % segments

		// Боковая грань (два треугольника)
		model.Triangles = append(model.Triangles, Triangle{
			V1: bottomVertices[idx1], V2: topVertices[idx1], V3: bottomVertices[idx2],
			UV1: texCoords[idx1*2], UV2: texCoords[idx1*2+1], UV3: texCoords[idx2*2],
			Normal: ecs.Vector3{X: bottomVertices[idx1].X * 2, Y: 0, Z: bottomVertices[idx1].Z * 2}.Normalize(),
		})

		model.Triangles = append(model.Triangles, Triangle{
			V1: bottomVertices[idx2], V2: topVertices[idx1], V3: topVertices[idx2],
			UV1: texCoords[idx2*2], UV2: texCoords[idx1*2+1], UV3: texCoords[idx2*2+1],
			Normal: ecs.Vector3{X: bottomVertices[idx2].X * 2, Y: 0, Z: bottomVertices[idx2].Z * 2}.Normalize(),
		})

		// Верхнее основание
		model.Triangles = append(model.Triangles, Triangle{
			V1: topCenter, V2: topVertices[idx1], V3: topVertices[idx2],
			UV1: TexCoord{X: 0.5, Y: 0.5}, UV2: TexCoord{X: 0.5 + math.Cos(float64(i)/float64(segments)*2*math.Pi)*0.5, Y: 0.5 + math.Sin(float64(i)/float64(segments)*2*math.Pi)*0.5}, UV3: TexCoord{X: 0.5 + math.Cos(float64(i+1)/float64(segments)*2*math.Pi)*0.5, Y: 0.5 + math.Sin(float64(i+1)/float64(segments)*2*math.Pi)*0.5},
			Normal: ecs.Vector3{X: 0, Y: 1, Z: 0},
		})

		// Нижнее основание
		model.Triangles = append(model.Triangles, Triangle{
			V1: bottomCenter, V2: bottomVertices[idx2], V3: bottomVertices[idx1],
			UV1: TexCoord{X: 0.5, Y: 0.5}, UV2: TexCoord{X: 0.5 + math.Cos(float64(i+1)/float64(segments)*2*math.Pi)*0.5, Y: 0.5 + math.Sin(float64(i+1)/float64(segments)*2*math.Pi)*0.5}, UV3: TexCoord{X: 0.5 + math.Cos(float64(i)/float64(segments)*2*math.Pi)*0.5, Y: 0.5 + math.Sin(float64(i)/float64(segments)*2*math.Pi)*0.5},
			Normal: ecs.Vector3{X: 0, Y: -1, Z: 0},
		})
	}

	// Устанавливаем bounding box
	model.BoundingBox.Min = Vertex{-0.5, -0.5, -0.5}
	model.BoundingBox.Max = Vertex{0.5, 0.5, 0.5}

	return model
}
