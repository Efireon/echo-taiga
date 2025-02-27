package biomes

import (
	"math"
	"math/rand"
)

// BiomeType represents the type of biome
type BiomeType string

const (
	BiomeTaiga       BiomeType = "taiga"
	BiomeMarsh       BiomeType = "marsh"
	BiomeRocky       BiomeType = "rocky"
	BiomeDenseTaiga  BiomeType = "dense_taiga"
	BiomeFrozenTaiga BiomeType = "frozen_taiga"
	BiomeDeadForest  BiomeType = "dead_forest"
	BiomeDistorted   BiomeType = "distorted" // Special biome for metamorphosed areas
)

// Biome represents a specific ecosystem within the world
type Biome struct {
	Type                  BiomeType
	TreeDensity           float64    // 0.0-1.0
	UndergrowthDensity    float64    // 0.0-1.0
	Humidity              float64    // 0.0-1.0
	Temperature           float64    // Range: -20.0 to 40.0 Celsius
	TerrainRoughness      float64    // 0.0-1.0
	BaseColorPalette      [3]float64 // RGB base color
	WaterContent          float64    // 0.0-1.0
	DangerLevel           float64    // 0.0-1.0
	AmbientSoundIntensity float64    // 0.0-1.0
	FogDensity            float64    // 0.0-1.0

	// Resources and features
	ResourceDistribution map[string]float64 // Resource type -> abundance
	SpecialFeatures      map[string]float64 // Feature -> probability
	FaunaTypes           []string
	FloraTypes           []string
}

// BiomeManager handles biome generation and transitions
type BiomeManager struct {
	Biomes           map[BiomeType]*Biome
	BiomeTransitions map[BiomeType]map[BiomeType]float64
	rng              *rand.Rand
}

// NewBiomeManager creates a new BiomeManager
func NewBiomeManager(seed int64) *BiomeManager {
	rng := rand.New(rand.NewSource(seed))

	bm := &BiomeManager{
		Biomes:           make(map[BiomeType]*Biome),
		BiomeTransitions: make(map[BiomeType]map[BiomeType]float64),
		rng:              rng,
	}

	// Initialize default biomes
	bm.initializeDefaultBiomes()

	// Set up transitions between biomes
	bm.initializeTransitions()

	return bm
}

// initializeDefaultBiomes creates the standard biomes
func (bm *BiomeManager) initializeDefaultBiomes() {
	// Standard Taiga
	bm.Biomes[BiomeTaiga] = &Biome{
		Type:                  BiomeTaiga,
		TreeDensity:           0.7,
		UndergrowthDensity:    0.5,
		Humidity:              0.6,
		Temperature:           5.0,
		TerrainRoughness:      0.5,
		BaseColorPalette:      [3]float64{0.2, 0.4, 0.2}, // Dark green
		WaterContent:          0.4,
		DangerLevel:           0.3,
		AmbientSoundIntensity: 0.5,
		FogDensity:            0.2,
		ResourceDistribution: map[string]float64{
			"wood":      0.8,
			"berries":   0.5,
			"mushrooms": 0.6,
			"animals":   0.5,
			"stones":    0.4,
			"herbs":     0.3,
		},
		SpecialFeatures: map[string]float64{
			"clearing":      0.2,
			"rocky_outcrop": 0.1,
			"fallen_tree":   0.3,
			"cave":          0.05,
		},
		FaunaTypes: []string{"wolf", "deer", "rabbit", "fox", "bear"},
		FloraTypes: []string{"pine", "spruce", "birch", "fern", "moss", "berry_bush"},
	}

	// Marsh
	bm.Biomes[BiomeMarsh] = &Biome{
		Type:                  BiomeMarsh,
		TreeDensity:           0.3,
		UndergrowthDensity:    0.8,
		Humidity:              0.9,
		Temperature:           8.0,
		TerrainRoughness:      0.2,
		BaseColorPalette:      [3]float64{0.3, 0.4, 0.1}, // Muddy green
		WaterContent:          0.8,
		DangerLevel:           0.6,
		AmbientSoundIntensity: 0.7,
		FogDensity:            0.6,
		ResourceDistribution: map[string]float64{
			"wood":      0.3,
			"berries":   0.2,
			"mushrooms": 0.8,
			"animals":   0.4,
			"stones":    0.1,
			"herbs":     0.7,
			"reeds":     0.9,
		},
		SpecialFeatures: map[string]float64{
			"pond":      0.7,
			"quicksand": 0.2,
			"fog_patch": 0.5,
			"dead_tree": 0.4,
		},
		FaunaTypes: []string{"frog", "snake", "bird", "insect", "fish"},
		FloraTypes: []string{"reeds", "willow", "lilypads", "moss", "fungi", "cattail"},
	}

	// Rocky
	bm.Biomes[BiomeRocky] = &Biome{
		Type:                  BiomeRocky,
		TreeDensity:           0.2,
		UndergrowthDensity:    0.3,
		Humidity:              0.3,
		Temperature:           2.0,
		TerrainRoughness:      0.9,
		BaseColorPalette:      [3]float64{0.5, 0.5, 0.5}, // Gray
		WaterContent:          0.2,
		DangerLevel:           0.5,
		AmbientSoundIntensity: 0.3,
		FogDensity:            0.1,
		ResourceDistribution: map[string]float64{
			"wood":      0.2,
			"berries":   0.1,
			"mushrooms": 0.2,
			"animals":   0.2,
			"stones":    0.9,
			"ores":      0.4,
			"herbs":     0.2,
		},
		SpecialFeatures: map[string]float64{
			"cave":       0.4,
			"cliff":      0.6,
			"rock_slide": 0.3,
			"hot_spring": 0.1,
		},
		FaunaTypes: []string{"goat", "eagle", "wolf", "insect"},
		FloraTypes: []string{"pine", "moss", "lichen", "hardy_shrub", "mountain_flower"},
	}

	// Dense Taiga
	bm.Biomes[BiomeDenseTaiga] = &Biome{
		Type:                  BiomeDenseTaiga,
		TreeDensity:           0.9,
		UndergrowthDensity:    0.7,
		Humidity:              0.6,
		Temperature:           4.0,
		TerrainRoughness:      0.6,
		BaseColorPalette:      [3]float64{0.1, 0.3, 0.1}, // Very dark green
		WaterContent:          0.5,
		DangerLevel:           0.6,
		AmbientSoundIntensity: 0.4,
		FogDensity:            0.5,
		ResourceDistribution: map[string]float64{
			"wood":      0.9,
			"berries":   0.4,
			"mushrooms": 0.7,
			"animals":   0.5,
			"stones":    0.3,
			"herbs":     0.4,
		},
		SpecialFeatures: map[string]float64{
			"ancient_tree":  0.3,
			"wolf_den":      0.2,
			"forest_shrine": 0.1,
			"fallen_tree":   0.4,
			"mushroom_ring": 0.2,
		},
		FaunaTypes: []string{"wolf", "bear", "deer", "owl", "squirrel"},
		FloraTypes: []string{"old_pine", "spruce", "fir", "moss", "fern", "mushroom"},
	}

	// Frozen Taiga
	bm.Biomes[BiomeFrozenTaiga] = &Biome{
		Type:                  BiomeFrozenTaiga,
		TreeDensity:           0.5,
		UndergrowthDensity:    0.2,
		Humidity:              0.3,
		Temperature:           -8.0,
		TerrainRoughness:      0.5,
		BaseColorPalette:      [3]float64{0.7, 0.8, 0.9}, // Pale blue-white
		WaterContent:          0.3,
		DangerLevel:           0.7,
		AmbientSoundIntensity: 0.2,
		FogDensity:            0.4,
		ResourceDistribution: map[string]float64{
			"wood":      0.6,
			"berries":   0.1,
			"mushrooms": 0.2,
			"animals":   0.3,
			"stones":    0.5,
			"herbs":     0.1,
			"ice":       0.8,
		},
		SpecialFeatures: map[string]float64{
			"frozen_lake": 0.3,
			"ice_cave":    0.2,
			"snow_drift":  0.6,
			"aurora":      0.1,
		},
		FaunaTypes: []string{"wolf", "fox", "rabbit", "owl", "elk"},
		FloraTypes: []string{"snow_pine", "dead_tree", "frost_flower", "winter_berry"},
	}

	// Dead Forest
	bm.Biomes[BiomeDeadForest] = &Biome{
		Type:                  BiomeDeadForest,
		TreeDensity:           0.6,
		UndergrowthDensity:    0.1,
		Humidity:              0.2,
		Temperature:           3.0,
		TerrainRoughness:      0.4,
		BaseColorPalette:      [3]float64{0.4, 0.3, 0.2}, // Brown
		WaterContent:          0.1,
		DangerLevel:           0.8,
		AmbientSoundIntensity: 0.2,
		FogDensity:            0.3,
		ResourceDistribution: map[string]float64{
			"wood":      0.7,
			"berries":   0.0,
			"mushrooms": 0.4,
			"animals":   0.2,
			"stones":    0.5,
			"herbs":     0.1,
			"bones":     0.6,
		},
		SpecialFeatures: map[string]float64{
			"bone_pile":    0.3,
			"charred_tree": 0.5,
			"ash_pile":     0.4,
			"carcass":      0.2,
		},
		FaunaTypes: []string{"raven", "wolf", "rat", "vulture"},
		FloraTypes: []string{"dead_pine", "dead_oak", "withered_bush", "strange_fungus"},
	}

	// Distorted (special biome for metamorphosed areas)
	bm.Biomes[BiomeDistorted] = &Biome{
		Type:                  BiomeDistorted,
		TreeDensity:           0.4,
		UndergrowthDensity:    0.6,
		Humidity:              0.5,
		Temperature:           10.0,
		TerrainRoughness:      0.8,
		BaseColorPalette:      [3]float64{0.5, 0.2, 0.5}, // Purple tint
		WaterContent:          0.5,
		DangerLevel:           0.9,
		AmbientSoundIntensity: 0.8,
		FogDensity:            0.7,
		ResourceDistribution: map[string]float64{
			"wood":      0.4,
			"berries":   0.3,
			"mushrooms": 0.9,
			"animals":   0.3,
			"stones":    0.4,
			"herbs":     0.2,
			"crystals":  0.6,
			"anomalous": 0.7,
		},
		SpecialFeatures: map[string]float64{
			"floating_rocks": 0.4,
			"gravity_well":   0.2,
			"reality_tear":   0.3,
			"strange_growth": 0.6,
			"color_shift":    0.8,
			"temporal_eddy":  0.3,
		},
		FaunaTypes: []string{"mutant_wolf", "distorted_deer", "shadow_creature", "floating_entity"},
		FloraTypes: []string{"twisted_pine", "glowing_mushroom", "floating_moss", "pulsating_flower"},
	}
}

// initializeTransitions sets up the probability of transitioning between biomes
func (bm *BiomeManager) initializeTransitions() {
	// Initialize transition maps for each biome
	for biomeType := range bm.Biomes {
		bm.BiomeTransitions[biomeType] = make(map[BiomeType]float64)
	}

	// Define transitions (probability of biome B appearing next to biome A)
	// Regular Taiga transitions
	bm.BiomeTransitions[BiomeTaiga][BiomeDenseTaiga] = 0.4
	bm.BiomeTransitions[BiomeTaiga][BiomeFrozenTaiga] = 0.2
	bm.BiomeTransitions[BiomeTaiga][BiomeMarsh] = 0.2
	bm.BiomeTransitions[BiomeTaiga][BiomeRocky] = 0.15
	bm.BiomeTransitions[BiomeTaiga][BiomeDeadForest] = 0.05

	// Dense Taiga transitions
	bm.BiomeTransitions[BiomeDenseTaiga][BiomeTaiga] = 0.5
	bm.BiomeTransitions[BiomeDenseTaiga][BiomeFrozenTaiga] = 0.2
	bm.BiomeTransitions[BiomeDenseTaiga][BiomeDeadForest] = 0.1
	bm.BiomeTransitions[BiomeDenseTaiga][BiomeMarsh] = 0.1
	bm.BiomeTransitions[BiomeDenseTaiga][BiomeRocky] = 0.1

	// Frozen Taiga transitions
	bm.BiomeTransitions[BiomeFrozenTaiga][BiomeTaiga] = 0.3
	bm.BiomeTransitions[BiomeFrozenTaiga][BiomeDenseTaiga] = 0.2
	bm.BiomeTransitions[BiomeFrozenTaiga][BiomeRocky] = 0.4
	bm.BiomeTransitions[BiomeFrozenTaiga][BiomeDeadForest] = 0.1

	// Marsh transitions
	bm.BiomeTransitions[BiomeMarsh][BiomeTaiga] = 0.4
	bm.BiomeTransitions[BiomeMarsh][BiomeDenseTaiga] = 0.2
	bm.BiomeTransitions[BiomeMarsh][BiomeDeadForest] = 0.2
	bm.BiomeTransitions[BiomeMarsh][BiomeRocky] = 0.1
	bm.BiomeTransitions[BiomeMarsh][BiomeFrozenTaiga] = 0.1

	// Rocky transitions
	bm.BiomeTransitions[BiomeRocky][BiomeTaiga] = 0.3
	bm.BiomeTransitions[BiomeRocky][BiomeFrozenTaiga] = 0.3
	bm.BiomeTransitions[BiomeRocky][BiomeDeadForest] = 0.2
	bm.BiomeTransitions[BiomeRocky][BiomeDenseTaiga] = 0.1
	bm.BiomeTransitions[BiomeRocky][BiomeMarsh] = 0.1

	// Dead Forest transitions
	bm.BiomeTransitions[BiomeDeadForest][BiomeTaiga] = 0.3
	bm.BiomeTransitions[BiomeDeadForest][BiomeRocky] = 0.3
	bm.BiomeTransitions[BiomeDeadForest][BiomeMarsh] = 0.2
	bm.BiomeTransitions[BiomeDeadForest][BiomeDenseTaiga] = 0.1
	bm.BiomeTransitions[BiomeDeadForest][BiomeFrozenTaiga] = 0.1

	// Distorted can transition to anything, and anything can (rarely) transition to Distorted
	for biomeType := range bm.Biomes {
		if biomeType != BiomeDistorted {
			bm.BiomeTransitions[BiomeDistorted][biomeType] = 0.15
			bm.BiomeTransitions[biomeType][BiomeDistorted] = 0.05
		}
	}
}

// GetBiomeAtPosition determines the biome at a given position based on noise values
func (bm *BiomeManager) GetBiomeAtPosition(x, y float64, noiseValues map[string]float64) BiomeType {
	// Extract relevant noise values
	temperature := noiseValues["temperature"] // -1.0 to 1.0
	humidity := noiseValues["humidity"]       // 0.0 to 1.0
	elevation := noiseValues["elevation"]     // 0.0 to 1.0
	anomaly := noiseValues["anomaly"]         // 0.0 to 1.0, higher means more likely to be distorted

	// Check for distorted biome first
	if anomaly > 0.7 {
		return BiomeDistorted
	}

	// Normalize temperature from -1.0,1.0 to 0.0,1.0
	temperature = (temperature + 1.0) / 2.0

	// Score each biome based on how well it matches the environment conditions
	biomeScores := make(map[BiomeType]float64)

	for biomeType, biome := range bm.Biomes {
		if biomeType == BiomeDistorted {
			continue // Skip distorted for normal selection
		}

		// Start with base score
		score := 1.0

		// Temperature match (0.0 = cold, 1.0 = hot)
		normalizedBiomeTemp := (biome.Temperature + 20.0) / 60.0 // Convert from -20,40 to 0,1
		tempDiff := math.Abs(normalizedBiomeTemp - temperature)
		tempFactor := 1.0 - tempDiff
		score *= tempFactor * 2.0 // Temperature is important

		// Humidity match
		humidityDiff := math.Abs(biome.Humidity - humidity)
		humidityFactor := 1.0 - humidityDiff
		score *= humidityFactor * 1.5

		// Elevation match - different biomes prefer different elevations
		var optimalElevation float64
		switch biomeType {
		case BiomeRocky:
			optimalElevation = 0.8
		case BiomeFrozenTaiga:
			optimalElevation = 0.7
		case BiomeTaiga, BiomeDenseTaiga:
			optimalElevation = 0.5
		case BiomeDeadForest:
			optimalElevation = 0.4
		case BiomeMarsh:
			optimalElevation = 0.2
		}

		elevationDiff := math.Abs(optimalElevation - elevation)
		elevationFactor := 1.0 - elevationDiff
		score *= elevationFactor * 1.8

		// Add a bit of randomness to avoid hard edges
		score *= 0.8 + 0.4*bm.rng.Float64()

		biomeScores[biomeType] = score
	}

	// Find the highest scoring biome
	var highestScore float64 = -1
	var selectedBiome BiomeType = BiomeTaiga // Default

	for biomeType, score := range biomeScores {
		if score > highestScore {
			highestScore = score
			selectedBiome = biomeType
		}
	}

	return selectedBiome
}

// GetNeighborBiome returns a suitable biome to place next to the given biome
func (bm *BiomeManager) GetNeighborBiome(currentBiome BiomeType) BiomeType {
	transitions, exists := bm.BiomeTransitions[currentBiome]
	if !exists {
		// If no transitions defined, return the same biome
		return currentBiome
	}

	// Sum all probabilities
	var totalProb float64 = 0
	for _, prob := range transitions {
		totalProb += prob
	}

	// If total probability is less than 1, add chance to stay the same
	if totalProb < 1.0 {
		transitions[currentBiome] = 1.0 - totalProb
		totalProb = 1.0
	}

	// Random roll
	roll := bm.rng.Float64() * totalProb
	var cumulativeProb float64 = 0

	for nextBiome, prob := range transitions {
		cumulativeProb += prob
		if roll <= cumulativeProb {
			return nextBiome
		}
	}

	// Default fallback
	return currentBiome
}

// GetBiome returns the biome data for a given biome type
func (bm *BiomeManager) GetBiome(biomeType BiomeType) *Biome {
	biome, exists := bm.Biomes[biomeType]
	if !exists {
		// Return default biome if not found
		return bm.Biomes[BiomeTaiga]
	}
	return biome
}

// CreateMetamorphosedBiome creates a modified version of a biome affected by a metamorphosis
func (bm *BiomeManager) CreateMetamorphosedBiome(baseBiome *Biome, metamorphosisOrder int, intensity float64) *Biome {
	// Create a copy of the base biome
	metamorphosed := &Biome{
		Type:                  baseBiome.Type,
		TreeDensity:           baseBiome.TreeDensity,
		UndergrowthDensity:    baseBiome.UndergrowthDensity,
		Humidity:              baseBiome.Humidity,
		Temperature:           baseBiome.Temperature,
		TerrainRoughness:      baseBiome.TerrainRoughness,
		BaseColorPalette:      baseBiome.BaseColorPalette,
		WaterContent:          baseBiome.WaterContent,
		DangerLevel:           baseBiome.DangerLevel,
		AmbientSoundIntensity: baseBiome.AmbientSoundIntensity,
		FogDensity:            baseBiome.FogDensity,
		ResourceDistribution:  make(map[string]float64),
		SpecialFeatures:       make(map[string]float64),
		FaunaTypes:            make([]string, len(baseBiome.FaunaTypes)),
		FloraTypes:            make([]string, len(baseBiome.FloraTypes)),
	}

	// Copy maps
	for k, v := range baseBiome.ResourceDistribution {
		metamorphosed.ResourceDistribution[k] = v
	}
	for k, v := range baseBiome.SpecialFeatures {
		metamorphosed.SpecialFeatures[k] = v
	}
	copy(metamorphosed.FaunaTypes, baseBiome.FaunaTypes)
	copy(metamorphosed.FloraTypes, baseBiome.FloraTypes)

	// Apply metamorphosis effects based on order
	switch metamorphosisOrder {
	case 1:
		// First order: Minor visual changes
		// Adjust color palette slightly
		colorShift := 0.1 * intensity
		metamorphosed.BaseColorPalette[0] += (bm.rng.Float64()*2 - 1) * colorShift
		metamorphosed.BaseColorPalette[1] += (bm.rng.Float64()*2 - 1) * colorShift
		metamorphosed.BaseColorPalette[2] += (bm.rng.Float64()*2 - 1) * colorShift

		// Adjust fog density
		metamorphosed.FogDensity += 0.1 * intensity

		// Adjust ambient sound
		metamorphosed.AmbientSoundIntensity += 0.1 * intensity

	case 2:
		// Second order: Structural changes
		// Adjust terrain and vegetation
		metamorphosed.TerrainRoughness += 0.2 * intensity
		metamorphosed.TreeDensity += (bm.rng.Float64()*2 - 1) * 0.2 * intensity
		metamorphosed.UndergrowthDensity += (bm.rng.Float64()*2 - 1) * 0.3 * intensity

		// More significant color shifts
		colorShift := 0.2 * intensity
		metamorphosed.BaseColorPalette[0] += (bm.rng.Float64()*2 - 1) * colorShift
		metamorphosed.BaseColorPalette[1] += (bm.rng.Float64()*2 - 1) * colorShift
		metamorphosed.BaseColorPalette[2] += (bm.rng.Float64()*2 - 1) * colorShift

		// Add strange features
		metamorphosed.SpecialFeatures["floating_objects"] = 0.2 * intensity
		metamorphosed.SpecialFeatures["strange_growth"] = 0.3 * intensity

		// Modify existing flora/fauna
		for i := range metamorphosed.FloraTypes {
			if bm.rng.Float64() < 0.3*intensity {
				metamorphosed.FloraTypes[i] = "altered_" + metamorphosed.FloraTypes[i]
			}
		}

	case 3:
		// Third order: Functional changes
		// More extreme modifications
		metamorphosed.TreeDensity = bm.rng.Float64()*intensity + (1-intensity)*baseBiome.TreeDensity
		metamorphosed.UndergrowthDensity = bm.rng.Float64()*intensity + (1-intensity)*baseBiome.UndergrowthDensity
		metamorphosed.TerrainRoughness = math.Min(1.0, baseBiome.TerrainRoughness+0.4*intensity)

		// Temperature and humidity shifts
		metamorphosed.Temperature += (bm.rng.Float64()*40 - 20) * intensity
		metamorphosed.Humidity = math.Max(0, math.Min(1, baseBiome.Humidity+(bm.rng.Float64()*2-1)*0.5*intensity))

		// Danger level increases
		metamorphosed.DangerLevel = math.Min(1.0, baseBiome.DangerLevel+0.3*intensity)

		// Add anomalous resources
		metamorphosed.ResourceDistribution["anomalous"] = 0.3 * intensity
		metamorphosed.ResourceDistribution["crystals"] = 0.2 * intensity

		// Add significant special features
		metamorphosed.SpecialFeatures["reality_tear"] = 0.2 * intensity
		metamorphosed.SpecialFeatures["gravity_well"] = 0.15 * intensity
		metamorphosed.SpecialFeatures["temporal_anomaly"] = 0.1 * intensity

		// Add mutated fauna
		metamorphosed.FaunaTypes = append(metamorphosed.FaunaTypes, "mutant_"+baseBiome.FaunaTypes[bm.rng.Intn(len(baseBiome.FaunaTypes))])

	case 4:
		// Fourth order: Systemic changes
		// Major environmental shifts
		metamorphosed.TreeDensity = bm.rng.Float64()
		metamorphosed.UndergrowthDensity = bm.rng.Float64()
		metamorphosed.Humidity = bm.rng.Float64()
		metamorphosed.Temperature = -20 + bm.rng.Float64()*60

		// Complete color palette shift
		metamorphosed.BaseColorPalette = [3]float64{
			bm.rng.Float64(),
			bm.rng.Float64(),
			bm.rng.Float64(),
		}

		// High danger
		metamorphosed.DangerLevel = 0.7 + 0.3*bm.rng.Float64()

		// Complete resource redistribution
		for k := range metamorphosed.ResourceDistribution {
			metamorphosed.ResourceDistribution[k] = bm.rng.Float64()
		}
		metamorphosed.ResourceDistribution["anomalous"] = 0.5 + 0.5*bm.rng.Float64()

		// Replace special features
		metamorphosed.SpecialFeatures = map[string]float64{
			"floating_rocks": 0.4 * intensity,
			"gravity_well":   0.3 * intensity,
			"reality_tear":   0.5 * intensity,
			"strange_growth": 0.6 * intensity,
			"color_shift":    0.7 * intensity,
			"temporal_eddy":  0.4 * intensity,
		}

		// Replace some fauna and flora with alien versions
		newFauna := make([]string, 0)
		for _, fauna := range baseBiome.FaunaTypes {
			if bm.rng.Float64() < 0.7*intensity {
				newFauna = append(newFauna, "distorted_"+fauna)
			} else {
				newFauna = append(newFauna, fauna)
			}
		}
		metamorphosed.FaunaTypes = append(newFauna, "shadow_creature", "floating_entity")

		newFlora := make([]string, 0)
		for _, flora := range baseBiome.FloraTypes {
			if bm.rng.Float64() < 0.7*intensity {
				newFlora = append(newFlora, "twisted_"+flora)
			} else {
				newFlora = append(newFlora, flora)
			}
		}
		metamorphosed.FloraTypes = append(newFlora, "glowing_mushroom", "pulsating_flower")

	case 5:
		// Fifth order: Fundamental changes
		// At this point, the biome is completely transformed
		// It essentially becomes the Distorted biome type
		distortedBiome := bm.Biomes[BiomeDistorted]

		metamorphosed.Type = BiomeDistorted
		metamorphosed.TreeDensity = distortedBiome.TreeDensity
		metamorphosed.UndergrowthDensity = distortedBiome.UndergrowthDensity
		metamorphosed.Humidity = distortedBiome.Humidity
		metamorphosed.Temperature = distortedBiome.Temperature
		metamorphosed.TerrainRoughness = distortedBiome.TerrainRoughness
		metamorphosed.BaseColorPalette = distortedBiome.BaseColorPalette
		metamorphosed.WaterContent = distortedBiome.WaterContent
		metamorphosed.DangerLevel = distortedBiome.DangerLevel
		metamorphosed.AmbientSoundIntensity = distortedBiome.AmbientSoundIntensity
		metamorphosed.FogDensity = distortedBiome.FogDensity

		// Tweak the distorted biome to make it unique
		metamorphosed.BaseColorPalette = [3]float64{
			0.3 + 0.7*bm.rng.Float64(),
			0.3 * bm.rng.Float64(),
			0.3 + 0.7*bm.rng.Float64(),
		}

		// Copy special features and resources from distorted biome
		metamorphosed.ResourceDistribution = make(map[string]float64)
		for k, v := range distortedBiome.ResourceDistribution {
			metamorphosed.ResourceDistribution[k] = v
		}

		metamorphosed.SpecialFeatures = make(map[string]float64)
		for k, v := range distortedBiome.SpecialFeatures {
			metamorphosed.SpecialFeatures[k] = v
		}

		// Use distorted fauna and flora
		metamorphosed.FaunaTypes = make([]string, len(distortedBiome.FaunaTypes))
		copy(metamorphosed.FaunaTypes, distortedBiome.FaunaTypes)

		metamorphosed.FloraTypes = make([]string, len(distortedBiome.FloraTypes))
		copy(metamorphosed.FloraTypes, distortedBiome.FloraTypes)

		// Add a few unique elements based on the original biome
		uniqueFeature := "remnant_" + string(baseBiome.Type)
		metamorphosed.SpecialFeatures[uniqueFeature] = 0.8
	}

	// Ensure values stay in valid ranges
	metamorphosed.TreeDensity = math.Max(0, math.Min(1, metamorphosed.TreeDensity))
	metamorphosed.UndergrowthDensity = math.Max(0, math.Min(1, metamorphosed.UndergrowthDensity))
	metamorphosed.Humidity = math.Max(0, math.Min(1, metamorphosed.Humidity))
	metamorphosed.TerrainRoughness = math.Max(0, math.Min(1, metamorphosed.TerrainRoughness))
	metamorphosed.WaterContent = math.Max(0, math.Min(1, metamorphosed.WaterContent))
	metamorphosed.DangerLevel = math.Max(0, math.Min(1, metamorphosed.DangerLevel))
	metamorphosed.AmbientSoundIntensity = math.Max(0, math.Min(1, metamorphosed.AmbientSoundIntensity))
	metamorphosed.FogDensity = math.Max(0, math.Min(1, metamorphosed.FogDensity))

	// Normalize color values
	for i := 0; i < 3; i++ {
		metamorphosed.BaseColorPalette[i] = math.Max(0, math.Min(1, metamorphosed.BaseColorPalette[i]))
	}

	return metamorphosed
}

// GenerateTreesForBiome returns a set of tree parameters appropriate for the biome
func (bm *BiomeManager) GenerateTreesForBiome(biomeType BiomeType, count int) []map[string]interface{} {
	biome := bm.GetBiome(biomeType)
	trees := make([]map[string]interface{}, count)

	for i := 0; i < count; i++ {
		treeType := biome.FloraTypes[bm.rng.Intn(len(biome.FloraTypes))]

		// Only process actual trees
		if !isTreeType(treeType) {
			continue
		}

		// Base parameters
		height := 0.0
		width := 0.0
		foliageDensity := 0.0
		color := [3]float64{0, 0, 0}

		// Set parameters based on tree type
		switch {
		case contains(treeType, "pine"):
			height = 8.0 + bm.rng.Float64()*6.0
			width = 1.0 + bm.rng.Float64()*2.0
			foliageDensity = 0.6 + bm.rng.Float64()*0.3
			color = [3]float64{0.0, 0.3 + bm.rng.Float64()*0.2, 0.0}

		case contains(treeType, "spruce"):
			height = 10.0 + bm.rng.Float64()*8.0
			width = 1.2 + bm.rng.Float64()*2.5
			foliageDensity = 0.7 + bm.rng.Float64()*0.3
			color = [3]float64{0.0, 0.2 + bm.rng.Float64()*0.2, 0.0}

		case contains(treeType, "birch"):
			height = 6.0 + bm.rng.Float64()*4.0
			width = 0.8 + bm.rng.Float64()*1.5
			foliageDensity = 0.5 + bm.rng.Float64()*0.3
			color = [3]float64{0.6, 0.8, 0.2 + bm.rng.Float64()*0.3}

		case contains(treeType, "willow"):
			height = 7.0 + bm.rng.Float64()*5.0
			width = 2.0 + bm.rng.Float64()*3.0
			foliageDensity = 0.8 + bm.rng.Float64()*0.2
			color = [3]float64{0.2, 0.4 + bm.rng.Float64()*0.2, 0.1}

		case contains(treeType, "dead") || contains(treeType, "withered"):
			height = 4.0 + bm.rng.Float64()*7.0
			width = 0.7 + bm.rng.Float64()*1.0
			foliageDensity = 0.0 + bm.rng.Float64()*0.2
			color = [3]float64{0.4, 0.3, 0.2}

		case contains(treeType, "twisted") || contains(treeType, "distorted"):
			height = 3.0 + bm.rng.Float64()*10.0
			width = 0.5 + bm.rng.Float64()*3.0
			foliageDensity = 0.2 + bm.rng.Float64()*0.7
			// Unusual colors for twisted trees
			color = [3]float64{
				0.3 + bm.rng.Float64()*0.7,
				0.0 + bm.rng.Float64()*0.5,
				0.3 + bm.rng.Float64()*0.7,
			}

		default:
			// Generic tree
			height = 5.0 + bm.rng.Float64()*5.0
			width = 1.0 + bm.rng.Float64()*2.0
			foliageDensity = 0.4 + bm.rng.Float64()*0.5
			color = [3]float64{0.1, 0.3 + bm.rng.Float64()*0.4, 0.1}
		}

		// Apply biome-specific modifiers
		// Adjust height by temperature
		if biome.Temperature < 0 {
			height *= 0.8 // Shorter in cold biomes
		} else if biome.Temperature > 20 {
			height *= 1.2 // Taller in warm biomes
		}

		// Adjust width by humidity
		width *= 0.7 + biome.Humidity*0.6

		// Adjust foliage by water content
		foliageDensity *= 0.6 + biome.WaterContent*0.8

		// Influence color by biome base color
		for i := 0; i < 3; i++ {
			color[i] = color[i]*0.8 + biome.BaseColorPalette[i]*0.2
		}

		// Vary trunk color
		trunkColor := [3]float64{0.3 + bm.rng.Float64()*0.2, 0.2 + bm.rng.Float64()*0.1, 0.1 + bm.rng.Float64()*0.1}

		// For twisted or distorted trees, make trunks unusual too
		if contains(treeType, "twisted") || contains(treeType, "distorted") {
			trunkColor = [3]float64{
				bm.rng.Float64() * 0.5,
				bm.rng.Float64() * 0.5,
				bm.rng.Float64() * 0.5,
			}
		}

		// Add special effects for certain tree types
		effects := make([]string, 0)
		if contains(treeType, "glowing") {
			effects = append(effects, "glow")
		}
		if contains(treeType, "twisted") {
			effects = append(effects, "distortion")
		}
		if contains(treeType, "floating") {
			effects = append(effects, "hover")
		}

		trees[i] = map[string]interface{}{
			"type":           treeType,
			"height":         height,
			"width":          width,
			"foliageDensity": foliageDensity,
			"foliageColor":   color,
			"trunkColor":     trunkColor,
			"health":         80.0 + bm.rng.Float64()*20.0,
			"effects":        effects,
		}
	}

	return trees
}

// Helper functions
func contains(s string, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

func isTreeType(floraType string) bool {
	treeTypes := []string{"pine", "spruce", "fir", "birch", "oak", "willow", "tree"}
	for _, treeType := range treeTypes {
		if contains(floraType, treeType) {
			return true
		}
	}
	return false
}
