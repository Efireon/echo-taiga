package symbols

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"echo-taiga/internal/engine/ecs"
)

// Symbol represents a mystical symbol that can be discovered and used in rituals
type Symbol struct {
	ID             string   // Unique identifier
	Name           string   // Name of the symbol
	Description    string   // Description of the symbol
	SymbolType     string   // Category: "elemental", "arcane", "primal", "void", etc.
	Complexity     float64  // 0-1: How complex/difficult the symbol is
	Power          float64  // 0-1: How powerful the symbol's effects are
	Meanings       []string // Conceptual meanings associated with the symbol
	RelatedSymbols []string // IDs of related symbols
	VisualID       string   // ID for the visual representation

	// Discovery information
	IsDiscovered      bool        // Whether the player has discovered this symbol
	DiscoveryTime     time.Time   // When the symbol was discovered
	DiscoveryLocation ecs.Vector3 // Where the symbol was discovered
	KnowledgeLevel    float64     // 0-1: Player's understanding of the symbol

	// Procedural generation parameters
	GenerationSeed int64   // Seed used to generate this symbol
	Distortion     float64 // 0-1: How distorted/corrupted the symbol is

	// Execution/gameplay effects
	RitualModifiers map[string]float64 // Modifiers when used in rituals
	WorldEffects    map[string]float64 // Direct effects on the world when observed
}

// Ritual represents a magical ritual that can be performed using symbols
type Ritual struct {
	ID          string // Unique identifier
	Name        string // Name of the ritual
	Description string // Description of the ritual

	// Components needed for the ritual
	RequiredSymbols  []string // Symbols required for the ritual
	RequiredItems    []string // Items required for the ritual
	RequiredLocation string   // Type of location needed (e.g., "water", "forest", "cave")

	// Execution information
	Actions      []string // Sequence of actions to perform the ritual
	Difficulty   float64  // 0-1: How difficult the ritual is to perform
	TimeRequired float64  // Seconds required to complete the ritual

	// Effects information
	Effects        []RitualEffect // Effects produced by the ritual
	SuccessChance  float64        // Base chance of success (0-1)
	FailureEffects []RitualEffect // Effects if ritual fails

	// Discovery and knowledge
	IsDiscovered   bool    // Whether the player has discovered this ritual
	KnowledgeLevel float64 // 0-1: Player's understanding of the ritual

	// Ritual progression information
	TimesPerformed  int       // How many times this ritual has been performed
	TimesSucceeded  int       // How many times this ritual has succeeded
	LastPerformTime time.Time // When the ritual was last performed

	// Procedural generation parameters
	GenerationSeed int64 // Seed used to generate this ritual

	// Evolution information
	EvolutionPath  []string // IDs of rituals that this can evolve into
	ParentRitual   string   // ID of the ritual this evolved from
	EvolutionLevel int      // How evolved this ritual is (0 = base ritual)
}

// RitualEffect represents an effect that a ritual can produce
type RitualEffect struct {
	Type        string   // Type of effect: "metamorphosis", "player", "item", "world", etc.
	Target      string   // Target of the effect: specific entity, area, player stat, etc.
	Value       float64  // Primary value/magnitude of the effect
	Duration    float64  // Duration in seconds (0 = permanent)
	Tags        []string // Tags describing the effect
	Description string   // Description of what happens

	// Spawning information (if effect spawns entities)
	SpawnEntityType string  // Type of entity to spawn
	SpawnCount      int     // Number of entities to spawn
	SpawnRadius     float64 // Radius in which to spawn entities

	// Metamorphosis information (if effect causes metamorphosis)
	MetamorphOrder int     // Order level of the metamorphosis
	MetamorphArea  float64 // Area of effect radius

	// Item information (if effect creates/modifies items)
	ItemID        string             // ID of the item to create/modify
	ItemModifiers map[string]float64 // Modifiers to apply to the item
}

// SymbolRegistry manages all symbols in the game
type SymbolRegistry struct {
	symbols           map[string]*Symbol // All symbols by ID
	discoveredSymbols map[string]*Symbol // Only discovered symbols

	// Organization by categories
	symbolsByType map[string][]*Symbol // Symbols organized by type

	// Procedural generation parameters
	baseSymbols    []*Symbol           // Template symbols for generation
	symbolPatterns []SymbolPattern     // Visual patterns for symbols
	meaningGroups  map[string][]string // Groups of related meanings

	// Loading/saving data
	savePath string // Path for saving/loading data

	mutex sync.RWMutex // Mutex for thread safety
}

// SymbolPattern represents a visual pattern for a symbol
type SymbolPattern struct {
	ID              string                 // Unique identifier
	BaseShape       string                 // Basic shape: "circle", "triangle", "square", etc.
	Elements        []SymbolElement        // Visual elements that make up the pattern
	Transformations []SymbolTransformation // Transformations applied to the pattern
	ColorScheme     []string               // Colors used in the pattern
	GenerationRules map[string]interface{} // Rules for procedural generation
}

// SymbolElement represents a visual element in a symbol
type SymbolElement struct {
	Type     string             // Type of element: "line", "arc", "dot", etc.
	Position [2]float64         // Position (normalized 0-1)
	Size     float64            // Size (normalized 0-1)
	Rotation float64            // Rotation in degrees
	Color    string             // Color reference from pattern's color scheme
	Params   map[string]float64 // Additional parameters specific to element type
}

// SymbolTransformation represents a transformation applied to a symbol pattern
type SymbolTransformation struct {
	Type   string             // Type of transformation: "rotate", "mirror", "repeat", etc.
	Params map[string]float64 // Parameters for the transformation
}

// RitualRegistry manages all rituals in the game
type RitualRegistry struct {
	rituals           map[string]*Ritual // All rituals by ID
	discoveredRituals map[string]*Ritual // Only discovered rituals

	// Organization by categories
	ritualsBySymbol map[string][]*Ritual // Rituals that use a particular symbol
	ritualsByEffect map[string][]*Ritual // Rituals that produce a particular effect type

	// Procedural generation parameters
	baseRituals     []*Ritual               // Template rituals for generation
	effectTemplates map[string]RitualEffect // Templates for ritual effects

	// Evolution tracking
	ritualEvolutionMap map[string][]string // Maps rituals to potential evolutions

	// Loading/saving data
	savePath string // Path for saving/loading data

	// Symbol reference (needed for ritual generation)
	symbolRegistry *SymbolRegistry

	mutex sync.RWMutex // Mutex for thread safety
}

// SymbolManager ties together the symbol and ritual systems
type SymbolManager struct {
	SymbolRegistry *SymbolRegistry // Registry of all symbols
	RitualRegistry *RitualRegistry // Registry of all rituals

	world *ecs.World // Reference to the ECS world

	// Tracking the player's interaction with the system
	playerKnowledge map[string]float64 // Knowledge level for each discovered symbol/ritual
	lastRitualCheck time.Time          // Time of last ritual check (for performance)

	// Callbacks for game events
	OnSymbolDiscovered func(symbol *Symbol)
	OnRitualDiscovered func(ritual *Ritual)
	OnRitualPerformed  func(ritual *Ritual, success bool, effects []RitualEffect)

	mutex sync.RWMutex // Mutex for thread safety
}

// NewSymbolRegistry creates a new symbol registry
func NewSymbolRegistry(savePath string) *SymbolRegistry {
	return &SymbolRegistry{
		symbols:           make(map[string]*Symbol),
		discoveredSymbols: make(map[string]*Symbol),
		symbolsByType:     make(map[string][]*Symbol),
		baseSymbols:       make([]*Symbol, 0),
		symbolPatterns:    make([]SymbolPattern, 0),
		meaningGroups:     make(map[string][]string),
		savePath:          savePath,
	}
}

// NewRitualRegistry creates a new ritual registry
func NewRitualRegistry(savePath string, symbolRegistry *SymbolRegistry) *RitualRegistry {
	return &RitualRegistry{
		rituals:            make(map[string]*Ritual),
		discoveredRituals:  make(map[string]*Ritual),
		ritualsBySymbol:    make(map[string][]*Ritual),
		ritualsByEffect:    make(map[string][]*Ritual),
		baseRituals:        make([]*Ritual, 0),
		effectTemplates:    make(map[string]RitualEffect),
		ritualEvolutionMap: make(map[string][]string),
		savePath:           savePath,
		symbolRegistry:     symbolRegistry,
	}
}

// NewSymbolManager creates a new symbol manager
func NewSymbolManager(world *ecs.World, savePath string) *SymbolManager {
	symbolRegistry := NewSymbolRegistry(filepath.Join(savePath, "symbols"))
	ritualRegistry := NewRitualRegistry(filepath.Join(savePath, "rituals"), symbolRegistry)

	return &SymbolManager{
		SymbolRegistry:  symbolRegistry,
		RitualRegistry:  ritualRegistry,
		world:           world,
		playerKnowledge: make(map[string]float64),
		lastRitualCheck: time.Now(),
	}
}

// Initialize initializes the symbol manager
func (sm *SymbolManager) Initialize() error {
	// Load base symbols and patterns
	err := sm.SymbolRegistry.LoadBaseSymbols()
	if err != nil {
		return fmt.Errorf("failed to load base symbols: %v", err)
	}

	// Load base rituals and effect templates
	err = sm.RitualRegistry.LoadBaseRituals()
	if err != nil {
		return fmt.Errorf("failed to load base rituals: %v", err)
	}

	// Try to load saved state
	err = sm.LoadState()
	if err != nil {
		// If there's no saved state, generate initial content
		sm.GenerateInitialContent()
	}

	// Register this as a system in the ECS world
	sm.world.AddSystem(sm)

	return nil
}

// LoadState loads the saved state of the symbol manager
func (sm *SymbolManager) LoadState() error {
	// Try to load symbols
	err := sm.SymbolRegistry.LoadState()
	if err != nil {
		return err
	}

	// Try to load rituals
	err = sm.RitualRegistry.LoadState()
	if err != nil {
		return err
	}

	// Load player knowledge
	knowledgePath := filepath.Join(sm.SymbolRegistry.savePath, "player_knowledge.json")
	if _, err := os.Stat(knowledgePath); os.IsNotExist(err) {
		return fmt.Errorf("player knowledge file does not exist")
	}

	data, err := ioutil.ReadFile(knowledgePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &sm.playerKnowledge)
	if err != nil {
		return err
	}

	return nil
}

// SaveState saves the current state of the symbol manager
func (sm *SymbolManager) SaveState() error {
	// Save symbols
	err := sm.SymbolRegistry.SaveState()
	if err != nil {
		return err
	}

	// Save rituals
	err = sm.RitualRegistry.SaveState()
	if err != nil {
		return err
	}

	// Save player knowledge
	knowledgePath := filepath.Join(sm.SymbolRegistry.savePath, "player_knowledge.json")

	// Make sure directory exists
	dir := filepath.Dir(knowledgePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Serialize and save
	data, err := json.MarshalIndent(sm.playerKnowledge, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(knowledgePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GenerateInitialContent generates the initial symbols and rituals
func (sm *SymbolManager) GenerateInitialContent() {
	// Generate symbols based on base templates
	for _, baseSymbol := range sm.SymbolRegistry.baseSymbols {
		for i := 0; i < 3; i++ { // Generate a few variations of each base symbol
			symbol := sm.GenerateSymbol(baseSymbol.SymbolType, i)
			sm.SymbolRegistry.AddSymbol(symbol)
		}
	}

	// Generate basic rituals that use these symbols
	for _, baseRitual := range sm.RitualRegistry.baseRituals {
		ritual := sm.GenerateRitual(baseRitual)
		sm.RitualRegistry.AddRitual(ritual)
	}
}

// GenerateSymbol creates a new procedurally generated symbol
func (sm *SymbolManager) GenerateSymbol(symbolType string, seed int) *Symbol {
	// Get a base symbol of the specified type as a template
	var baseSymbol *Symbol
	for _, s := range sm.SymbolRegistry.baseSymbols {
		if s.SymbolType == symbolType {
			baseSymbol = s
			break
		}
	}

	if baseSymbol == nil {
		// Fall back to any base symbol if the specified type isn't found
		if len(sm.SymbolRegistry.baseSymbols) > 0 {
			1
			baseSymbol = sm.SymbolRegistry.baseSymbols[0]
		} else {
			// Create a minimal base symbol if none exist
			baseSymbol = &Symbol{
				ID:          "default_symbol",
				Name:        "Unknown Symbol",
				Description: "A mysterious symbol of unknown origin",
				SymbolType:  "arcane",
				Complexity:  0.5,
				Power:       0.5,
				Meanings:    []string{"mystery", "unknown"},
				VisualID:    "default_symbol_visual",
			}
		}
	}

	// Generate a new symbol based on the template
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(seed)))

	// Create unique ID
	symbolID := fmt.Sprintf("%s_%s_%d", symbolType, generateRandomString(4), seed)

	// Create variation of the name
	nameAdjectives := []string{"Ancient", "Mystic", "Hidden", "Forgotten", "Primal", "Eldritch", "Secret", "Twisted", "Eternal", "Undying"}
	nameNouns := []string{"Sigil", "Rune", "Glyph", "Mark", "Sign", "Symbol", "Emblem", "Insignia", "Seal", "Inscription"}

	symbolName := fmt.Sprintf("%s %s of %s",
		nameAdjectives[r.Intn(len(nameAdjectives))],
		nameNouns[r.Intn(len(nameNouns))],
		generateSymbolNameSuffix(symbolType, r))

	// Create description
	descTemplates := []string{
		"A %s symbol that %s when observed closely. It seems to %s.",
		"This %s marking appears to %s. Those who study it report %s.",
		"A %s sigil that %s. Legend says it was %s.",
	}

	descTemplate := descTemplates[r.Intn(len(descTemplates))]
	symbolDesc := fmt.Sprintf(descTemplate,
		generateSymbolAdjective(symbolType, r),
		generateSymbolVerb(symbolType, r),
		generateSymbolEffect(symbolType, r))

	// Generate meanings based on symbol type
	meanings := generateSymbolMeanings(symbolType, r, 2+r.Intn(3))

	// Generate visual ID
	visualID := fmt.Sprintf("symbol_%s_%d", symbolType, seed)

	// Create the new symbol
	symbol := &Symbol{
		ID:             symbolID,
		Name:           symbolName,
		Description:    symbolDesc,
		SymbolType:     symbolType,
		Complexity:     0.3 + r.Float64()*0.7, // Between 0.3 and 1.0
		Power:          0.2 + r.Float64()*0.6, // Between 0.2 and 0.8
		Meanings:       meanings,
		RelatedSymbols: []string{},
		VisualID:       visualID,
		IsDiscovered:   false,
		KnowledgeLevel: 0.0,
		GenerationSeed: time.Now().UnixNano() + int64(seed),
		Distortion:     r.Float64() * 0.3, // Some small random distortion
		RitualModifiers: map[string]float64{
			"power":     0.8 + r.Float64()*0.4,
			"stability": 0.7 + r.Float64()*0.6,
			"duration":  0.9 + r.Float64()*0.2,
		},
		WorldEffects: map[string]float64{},
	}

	// Add some type-specific world effects
	switch symbolType {
	case "elemental":
		symbol.WorldEffects["elemental_resonance"] = 0.5 + r.Float64()*0.5
	case "arcane":
		symbol.WorldEffects["magic_amplification"] = 0.4 + r.Float64()*0.6
	case "primal":
		symbol.WorldEffects["nature_connection"] = 0.6 + r.Float64()*0.4
	case "void":
		symbol.WorldEffects["reality_distortion"] = 0.7 + r.Float64()*0.3
	}

	return symbol
}

// GenerateRitual creates a new procedurally generated ritual
func (sm *SymbolManager) GenerateRitual(baseRitual *Ritual) *Ritual {
	// Get available symbols to use in the ritual
	availableSymbols := sm.SymbolRegistry.GetSymbolsByType(baseRitual.RequiredLocation)
	if len(availableSymbols) == 0 {
		// Fall back to all symbols if none match the location
		availableSymbols = sm.SymbolRegistry.GetAllSymbols()
	}

	if len(availableSymbols) == 0 {
		// Can't generate a ritual without symbols
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Determine how many symbols to require (2-4)
	symbolCount := 2 + r.Intn(3)
	if symbolCount > len(availableSymbols) {
		symbolCount = len(availableSymbols)
	}

	// Select random symbols
	requiredSymbols := make([]string, 0, symbolCount)
	symbolIndices := make([]int, len(availableSymbols))
	for i := range symbolIndices {
		symbolIndices[i] = i
	}

	// Shuffle the indices
	for i := range symbolIndices {
		j := r.Intn(i + 1)
		symbolIndices[i], symbolIndices[j] = symbolIndices[j], symbolIndices[i]
	}

	// Take the first symbolCount indices
	for i := 0; i < symbolCount; i++ {
		requiredSymbols = append(requiredSymbols, availableSymbols[symbolIndices[i]].ID)
	}

	// Create unique ID
	ritualID := fmt.Sprintf("ritual_%s_%s", baseRitual.RequiredLocation, generateRandomString(6))

	// Create name
	nameAdjectives := []string{"Ritual", "Ceremony", "Rite", "Invocation", "Summoning", "Binding", "Banishing", "Awakening"}
	namePrefix := nameAdjectives[r.Intn(len(nameAdjectives))]

	// Base the name on the primary symbol
	primarySymbol := sm.SymbolRegistry.GetSymbol(requiredSymbols[0])
	ritualName := fmt.Sprintf("%s of %s", namePrefix, generateRitualNameSuffix(primarySymbol, r))

	// Generate description
	descTemplates := []string{
		"A %s ritual that %s when performed correctly. It requires %s.",
		"This %s ceremony must be performed %s. It is said to %s.",
		"An ancient rite that %s. It must be conducted %s with %s.",
	}

	locationDesc := generateLocationDescription(baseRitual.RequiredLocation)

	descTemplate := descTemplates[r.Intn(len(descTemplates))]
	ritualDesc := fmt.Sprintf(descTemplate,
		generateRitualAdjective(baseRitual.RequiredLocation, r),
		generateRitualEffect(baseRitual.RequiredLocation, r),
		locationDesc)

	// Generate required items
	itemCount := r.Intn(3) // 0 to 2 items
	requiredItems := make([]string, 0, itemCount)

	itemTypes := []string{"herb", "mineral", "bone", "fluid", "cloth", "tool"}
	for i := 0; i < itemCount; i++ {
		itemType := itemTypes[r.Intn(len(itemTypes))]
		item := generateRitualItem(itemType, r)
		requiredItems = append(requiredItems, item)
	}

	// Generate actions
	actionCount := 3 + r.Intn(3) // 3 to 5 actions
	actions := make([]string, 0, actionCount)

	actionTypes := []string{"draw", "place", "chant", "burn", "pour", "scatter", "meditate", "gesture"}
	for i := 0; i < actionCount; i++ {
		actionType := actionTypes[r.Intn(len(actionTypes))]
		action := generateRitualAction(actionType, r)
		actions = append(actions, action)
	}

	// Generate effects
	effectCount := 1 + r.Intn(2) // 1 to 2 effects
	effects := make([]RitualEffect, 0, effectCount)

	// Calculate total power of the symbols
	totalPower := 0.0
	for _, symbolID := range requiredSymbols {
		symbol := sm.SymbolRegistry.GetSymbol(symbolID)
		if symbol != nil {
			totalPower += symbol.Power
		}
	}
	averagePower := totalPower / float64(len(requiredSymbols))

	// Generate primary effect based on ritual location
	primaryEffect := generatePrimaryEffect(baseRitual.RequiredLocation, averagePower, r)
	effects = append(effects, primaryEffect)

	// Add additional effects if needed
	if effectCount > 1 {
		secondaryEffect := generateSecondaryEffect(baseRitual.RequiredLocation, averagePower*0.7, r)
		effects = append(effects, secondaryEffect)
	}

	// Generate failure effects
	failureEffects := make([]RitualEffect, 0, 1)
	failureEffect := generateFailureEffect(baseRitual.RequiredLocation, averagePower, r)
	failureEffects = append(failureEffects, failureEffect)

	// Create the ritual
	ritual := &Ritual{
		ID:               ritualID,
		Name:             ritualName,
		Description:      ritualDesc,
		RequiredSymbols:  requiredSymbols,
		RequiredItems:    requiredItems,
		RequiredLocation: baseRitual.RequiredLocation,
		Actions:          actions,
		Difficulty:       0.3 + r.Float64()*0.5, // Between 0.3 and 0.8
		TimeRequired:     30 + r.Float64()*120,  // Between 30 and 150 seconds
		Effects:          effects,
		SuccessChance:    0.6 + r.Float64()*0.3, // Between 0.6 and 0.9
		FailureEffects:   failureEffects,
		IsDiscovered:     false,
		KnowledgeLevel:   0.0,
		TimesPerformed:   0,
		TimesSucceeded:   0,
		GenerationSeed:   time.Now().UnixNano(),
		EvolutionPath:    []string{},
		ParentRitual:     "",
		EvolutionLevel:   0,
	}

	return ritual
}

// RequiredComponents returns the components required by this system
func (sm *SymbolManager) RequiredComponents() []ecs.ComponentID {
	return []ecs.ComponentID{ecs.SymbolComponentID}
}

// Update is called once per frame
func (sm *SymbolManager) Update(deltaTime float64) {
	// Check for nearby symbols that the player can discover
	// Only check every 0.5 seconds to avoid performance issues
	if time.Since(sm.lastRitualCheck) < 500*time.Millisecond {
		return
	}
	sm.lastRitualCheck = time.Now()

	// Get player entity
	playerEntities := sm.world.GetEntitiesWithTag("player")
	if len(playerEntities) == 0 {
		return
	}
	player := playerEntities[0]

	// Get player position
	playerTransform, has := player.GetComponent(ecs.TransformComponentID)
	if !has {
		return
	}
	playerPos := playerTransform.(*ecs.TransformComponent).Position

	// Check for symbols in the world
	symbolEntities := sm.world.GetEntitiesWithComponent(ecs.SymbolComponentID)
	for _, entity := range symbolEntities {
		symbolComp, _ := entity.GetComponent(ecs.SymbolComponentID)
		symbol := symbolComp.(*ecs.SymbolComponent)

		// Skip if already discovered
		if symbol.Discovered {
			continue
		}

		// Check if player is close enough to discover
		entityTransform, has := entity.GetComponent(ecs.TransformComponentID)
		if !has {
			continue
		}
		entityPos := entityTransform.(*ecs.TransformComponent).Position

		distance := playerPos.Distance(entityPos)
		if distance <= symbol.DiscoveryRadius {
			// Discover the symbol
			symbol.Discovered = true

			// Get the registered symbol
			regSymbol := sm.SymbolRegistry.GetSymbol(symbol.SymbolID)
			if regSymbol != nil {
				sm.DiscoverSymbol(regSymbol, playerPos)
			}
		}
	}

	// Update symbol knowledge levels
	sm.updateSymbolKnowledge(deltaTime)
}

// DiscoverSymbol marks a symbol as discovered by the player
func (sm *SymbolManager) DiscoverSymbol(symbol *Symbol, location ecs.Vector3) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if already discovered
	if symbol.IsDiscovered {
		return
	}

	// Mark as discovered
	symbol.IsDiscovered = true
	symbol.DiscoveryTime = time.Now()
	symbol.DiscoveryLocation = location
	symbol.KnowledgeLevel = 0.1 // Initial understanding

	// Add to discovered symbols
	sm.SymbolRegistry.discoveredSymbols[symbol.ID] = symbol

	// Initialize knowledge level
	sm.playerKnowledge[symbol.ID] = 0.1

	// Check for new ritual discoveries based on this symbol
	sm.checkForRitualDiscoveries()

	// Trigger callback if set
	if sm.OnSymbolDiscovered != nil {
		sm.OnSymbolDiscovered(symbol)
	}
}

// PerformRitual attempts to perform a ritual
func (sm *SymbolManager) PerformRitual(ritual *Ritual, location ecs.Vector3, items []string, playerSkill float64) (bool, []RitualEffect) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Update ritual stats
	ritual.TimesPerformed++
	ritual.LastPerformTime = time.Now()

	// Check if this is a valid location
	locationValid := checkRitualLocation(ritual.RequiredLocation, location, sm.world)

	// Check if all required items are present
	itemsValid := true
	for _, requiredItem := range ritual.RequiredItems {
		if !containsString(items, requiredItem) {
			itemsValid = false
			break
		}
	}

	// Calculate success chance
	successChance := ritual.SuccessChance

	// Location and items affect success
	if !locationValid {
		successChance *= 0.5
	}
	if !itemsValid {
		successChance *= 0.7
	}

	// Player skill affects success
	successChance *= (0.5 + 0.5*playerSkill)

	// Player knowledge of the ritual affects success
	knowledgeLevel := sm.GetKnowledgeLevel(ritual.ID)
	successChance *= (0.5 + 0.5*knowledgeLevel)

	// Random factor
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Float64()

	// Check for success
	success := roll < successChance

	var effects []RitualEffect
	if success {
		// Ritual succeeded
		ritual.TimesSucceeded++
		effects = ritual.Effects

		// Increase knowledge
		sm.IncreaseKnowledge(ritual.ID, 0.1)

		// Check for ritual evolution
		if ritual.TimesSucceeded >= 3 && len(ritual.EvolutionPath) > 0 {
			sm.EvolveRitual(ritual)
		}
	} else {
		// Ritual failed
		effects = ritual.FailureEffects

		// Still gain some knowledge
		sm.IncreaseKnowledge(ritual.ID, 0.05)
	}

	// Trigger callback if set
	if sm.OnRitualPerformed != nil {
		sm.OnRitualPerformed(ritual, success, effects)
	}

	return success, effects
}

// GetKnowledgeLevel returns the player's knowledge level for a symbol or ritual
func (sm *SymbolManager) GetKnowledgeLevel(id string) float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	level, exists := sm.playerKnowledge[id]
	if !exists {
		return 0.0
	}
	return level
}

// IncreaseKnowledge increases the player's knowledge of a symbol or ritual
func (sm *SymbolManager) IncreaseKnowledge(id string, amount float64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Get current knowledge
	currentLevel := sm.playerKnowledge[id]

	// Increase knowledge with diminishing returns
	newLevel := currentLevel + amount*(1.0-currentLevel*0.8)
	if newLevel > 1.0 {
		newLevel = 1.0
	}

	// Update knowledge
	sm.playerKnowledge[id] = newLevel

	// Update symbol or ritual
	if symbol := sm.SymbolRegistry.GetSymbol(id); symbol != nil {
		symbol.KnowledgeLevel = newLevel
	} else if ritual := sm.RitualRegistry.GetRitual(id); ritual != nil {
		ritual.KnowledgeLevel = newLevel
	}

	// Check for potential ritual discoveries
	sm.checkForRitualDiscoveries()
}

// EvolveRitual evolves a ritual into a more advanced form
func (sm *SymbolManager) EvolveRitual(ritual *Ritual) {
	// Check if ritual can evolve
	if len(ritual.EvolutionPath) == 0 {
		return
	}

	// Get the next evolution
	evolutionID := ritual.EvolutionPath[0]
	evolvedRitual := sm.RitualRegistry.GetRitual(evolutionID)

	if evolvedRitual == nil {
		// Evolution doesn't exist yet, create it
		evolvedRitual = sm.GenerateEvolvedRitual(ritual)
		if evolvedRitual == nil {
			return
		}

		// Add to registry
		sm.RitualRegistry.AddRitual(evolvedRitual)
	}

	// Mark evolved ritual as discovered
	if !evolvedRitual.IsDiscovered {
		evolvedRitual.IsDiscovered = true
		sm.RitualRegistry.discoveredRituals[evolvedRitual.ID] = evolvedRitual

		// Initial knowledge
		evolvedRitual.KnowledgeLevel = 0.2
		sm.playerKnowledge[evolvedRitual.ID] = 0.2

		// Trigger callback if set
		if sm.OnRitualDiscovered != nil {
			sm.OnRitualDiscovered(evolvedRitual)
		}
	}
}

// GenerateEvolvedRitual creates an evolved version of a ritual
func (sm *SymbolManager) GenerateEvolvedRitual(baseRitual *Ritual) *Ritual {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create unique ID
	evolvedID := fmt.Sprintf("%s_evolved_%s", baseRitual.ID, generateRandomString(4))

	// Create evolved name
	evolvedAdjectives := []string{"Advanced", "Greater", "Empowered", "Ascended", "Refined", "Mastered"}
	adjective := evolvedAdjectives[r.Intn(len(evolvedAdjectives))]

	// Strip any existing evolved adjectives from the name
	baseName := baseRitual.Name
	for _, a := range evolvedAdjectives {
		baseName = strings.TrimPrefix(baseName, a+" ")
	}

	evolvedName := fmt.Sprintf("%s %s", adjective, baseName)

	// Create evolved description
	evolvedDesc := fmt.Sprintf("An evolved form of the %s. %s", baseRitual.Name, baseRitual.Description)

	// Use the same symbols, but possibly add one more
	evolvedSymbols := make([]string, len(baseRitual.RequiredSymbols))
	copy(evolvedSymbols, baseRitual.RequiredSymbols)

	// 50% chance to add another symbol
	if r.Float64() < 0.5 {
		availableSymbols := sm.SymbolRegistry.GetAllSymbols()
		if len(availableSymbols) > len(evolvedSymbols) {
			// Find a symbol that's not already included
			for _, symbol := range availableSymbols {
				if !containsString(evolvedSymbols, symbol.ID) {
					evolvedSymbols = append(evolvedSymbols, symbol.ID)
					break
				}
			}
		}
	}

	// Keep the same items, but might add one more
	evolvedItems := make([]string, len(baseRitual.RequiredItems))
	copy(evolvedItems, baseRitual.RequiredItems)

	// 30% chance to add another item
	if r.Float64() < 0.3 {
		itemTypes := []string{"herb", "mineral", "bone", "fluid", "cloth", "tool"}
		itemType := itemTypes[r.Intn(len(itemTypes))]
		newItem := generateRitualItem(itemType, r)

		// Make sure it's not a duplicate
		if !containsString(evolvedItems, newItem) {
			evolvedItems = append(evolvedItems, newItem)
		}
	}

	// Keep the same actions, but might add one more or modify existing
	evolvedActions := make([]string, len(baseRitual.Actions))
	copy(evolvedActions, baseRitual.Actions)

	// 40% chance to add another action
	if r.Float64() < 0.4 {
		actionTypes := []string{"draw", "place", "chant", "burn", "pour", "scatter", "meditate", "gesture"}
		actionType := actionTypes[r.Intn(len(actionTypes))]
		newAction := generateRitualAction(actionType, r)
		evolvedActions = append(evolvedActions, newAction)
	}

	// Enhanced effects
	evolvedEffects := make([]RitualEffect, len(baseRitual.Effects))
	for i, effect := range baseRitual.Effects {
		// Create enhanced copy of the effect
		enhancedEffect := effect
		enhancedEffect.Value *= 1.3 + r.Float64()*0.3    // 1.3x to 1.6x stronger
		enhancedEffect.Duration *= 1.5 + r.Float64()*0.5 // 1.5x to 2.0x longer

		evolvedEffects[i] = enhancedEffect
	}

	// 30% chance to add another effect
	if r.Float64() < 0.3 {
		// Calculate average power of the symbols
		totalPower := 0.0
		for _, symbolID := range evolvedSymbols {
			symbol := sm.SymbolRegistry.GetSymbol(symbolID)
			if symbol != nil {
				totalPower += symbol.Power
			}
		}
		averagePower := totalPower / float64(len(evolvedSymbols))

		newEffect := generateSecondaryEffect(baseRitual.RequiredLocation, averagePower, r)
		evolvedEffects = append(evolvedEffects, newEffect)
	}

	// Enhanced failure effects
	evolvedFailureEffects := make([]RitualEffect, len(baseRitual.FailureEffects))
	for i, effect := range baseRitual.FailureEffects {
		// Create enhanced copy of the effect
		enhancedEffect := effect
		enhancedEffect.Value *= 1.2 // Slightly more dangerous

		evolvedFailureEffects[i] = enhancedEffect
	}

	// Create the evolved ritual
	evolvedRitual := &Ritual{
		ID:               evolvedID,
		Name:             evolvedName,
		Description:      evolvedDesc,
		RequiredSymbols:  evolvedSymbols,
		RequiredItems:    evolvedItems,
		RequiredLocation: baseRitual.RequiredLocation,
		Actions:          evolvedActions,
		Difficulty:       baseRitual.Difficulty * 1.2,   // More difficult
		TimeRequired:     baseRitual.TimeRequired * 1.5, // Takes longer
		Effects:          evolvedEffects,
		SuccessChance:    baseRitual.SuccessChance * 0.9, // Slightly lower chance
		FailureEffects:   evolvedFailureEffects,
		IsDiscovered:     false,
		KnowledgeLevel:   0,
		TimesPerformed:   0,
		TimesSucceeded:   0,
		GenerationSeed:   time.Now().UnixNano(),
		EvolutionPath:    []string{},
		ParentRitual:     baseRitual.ID,
		EvolutionLevel:   baseRitual.EvolutionLevel + 1,
	}

	return evolvedRitual
}

// checkForRitualDiscoveries checks if the player has discovered enough symbols to learn new rituals
func (sm *SymbolManager) checkForRitualDiscoveries() {
	// Get all undiscovered rituals
	undiscoveredRituals := sm.RitualRegistry.GetUndiscoveredRituals()

	for _, ritual := range undiscoveredRituals {
		// Check if the player knows enough about the required symbols
		canDiscover := true
		totalKnowledge := 0.0

		for _, symbolID := range ritual.RequiredSymbols {
			symbol := sm.SymbolRegistry.GetSymbol(symbolID)
			if symbol == nil || !symbol.IsDiscovered {
				canDiscover = false
				break
			}

			knowledge := sm.GetKnowledgeLevel(symbolID)
			totalKnowledge += knowledge
		}

		// Need an average knowledge of at least 0.4 to discover the ritual
		if canDiscover && totalKnowledge/float64(len(ritual.RequiredSymbols)) >= 0.4 {
			// Discover the ritual
			ritual.IsDiscovered = true
			sm.RitualRegistry.discoveredRituals[ritual.ID] = ritual

			// Initial knowledge
			ritual.KnowledgeLevel = 0.1
			sm.playerKnowledge[ritual.ID] = 0.1

			// Trigger callback if set
			if sm.OnRitualDiscovered != nil {
				sm.OnRitualDiscovered(ritual)
			}
		}
	}
}

// updateSymbolKnowledge updates the knowledge levels based on study and usage
func (sm *SymbolManager) updateSymbolKnowledge(deltaTime float64) {
	// Get all discovered symbols
	discoveredSymbols := sm.SymbolRegistry.GetDiscoveredSymbols()

	// Get all discovered rituals
	discoveredRituals := sm.RitualRegistry.GetDiscoveredRituals()

	// Update knowledge based on complexity and time
	// This simulates the player gradually learning more about symbols they've discovered
	for _, symbol := range discoveredSymbols {
		// Knowledge increases more slowly for complex symbols
		knowledgeGain := (0.0001 * deltaTime) / (symbol.Complexity * 2)
		sm.IncreaseKnowledge(symbol.ID, knowledgeGain)
	}

	// For rituals, knowledge only increases through performing them
	// But still allow a tiny gain over time to represent the player thinking about them
	for _, ritual := range discoveredRituals {
		// Knowledge increases very slowly for difficult rituals
		knowledgeGain := (0.00005 * deltaTime) / (ritual.Difficulty * 2)
		sm.IncreaseKnowledge(ritual.ID, knowledgeGain)
	}
}

// Helper functions for symbol registry

// LoadBaseSymbols loads base symbols from files
func (sr *SymbolRegistry) LoadBaseSymbols() error {
	basePath := filepath.Join(sr.savePath, "base")

	// Create directory if it doesn't exist
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err = os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			return err
		}

		// Create default base symbols
		err = sr.CreateDefaultBaseSymbols(basePath)
		if err != nil {
			return err
		}
	}

	// Load files
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".json" {
			// Load symbol
			filePath := filepath.Join(basePath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			var symbol Symbol
			err = json.Unmarshal(data, &symbol)
			if err != nil {
				continue
			}

			// Add to base symbols
			sr.baseSymbols = append(sr.baseSymbols, &symbol)
		}
	}

	// Load patterns
	patternsPath := filepath.Join(sr.savePath, "patterns")
	if _, err := os.Stat(patternsPath); os.IsNotExist(err) {
		err = os.MkdirAll(patternsPath, os.ModePerm)
		if err != nil {
			return err
		}

		// Create default patterns
		err = sr.CreateDefaultPatterns(patternsPath)
		if err != nil {
			return err
		}
	}

	// Load pattern files
	files, err = ioutil.ReadDir(patternsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".json" {
			// Load pattern
			filePath := filepath.Join(patternsPath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			var pattern SymbolPattern
			err = json.Unmarshal(data, &pattern)
			if err != nil {
				continue
			}

			// Add to patterns
			sr.symbolPatterns = append(sr.symbolPatterns, pattern)
		}
	}

	// Load meaning groups
	meaningsPath := filepath.Join(sr.savePath, "meanings.json")
	if _, err := os.Stat(meaningsPath); os.IsNotExist(err) {
		// Create default meanings
		err = sr.CreateDefaultMeanings(meaningsPath)
		if err != nil {
			return err
		}
	}

	// Load meanings
	data, err := ioutil.ReadFile(meaningsPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &sr.meaningGroups)
	if err != nil {
		return err
	}

	return nil
}

// CreateDefaultBaseSymbols creates default base symbols
func (sr *SymbolRegistry) CreateDefaultBaseSymbols(basePath string) error {
	// Create base symbols for different types
	baseSymbols := []Symbol{
		{
			ID:          "base_elemental",
			Name:        "Elemental Symbol",
			Description: "A symbol representing elemental forces",
			SymbolType:  "elemental",
			Complexity:  0.5,
			Power:       0.6,
			Meanings:    []string{"elements", "nature", "force"},
			VisualID:    "base_elemental_visual",
			RitualModifiers: map[string]float64{
				"power":     1.2,
				"stability": 0.8,
				"duration":  1.0,
			},
			WorldEffects: map[string]float64{
				"elemental_resonance": 1.0,
			},
		},
		{
			ID:          "base_arcane",
			Name:        "Arcane Symbol",
			Description: "A symbol of arcane power",
			SymbolType:  "arcane",
			Complexity:  0.7,
			Power:       0.7,
			Meanings:    []string{"magic", "knowledge", "power"},
			VisualID:    "base_arcane_visual",
			RitualModifiers: map[string]float64{
				"power":     1.3,
				"stability": 0.7,
				"duration":  1.2,
			},
			WorldEffects: map[string]float64{
				"magic_amplification": 1.0,
			},
		},
		{
			ID:          "base_primal",
			Name:        "Primal Symbol",
			Description: "A symbol connected to primal forces",
			SymbolType:  "primal",
			Complexity:  0.4,
			Power:       0.8,
			Meanings:    []string{"life", "death", "growth"},
			VisualID:    "base_primal_visual",
			RitualModifiers: map[string]float64{
				"power":     1.1,
				"stability": 1.0,
				"duration":  0.9,
			},
			WorldEffects: map[string]float64{
				"nature_connection": 1.0,
			},
		},
		{
			ID:          "base_void",
			Name:        "Void Symbol",
			Description: "A symbol representing the void between realities",
			SymbolType:  "void",
			Complexity:  0.9,
			Power:       0.9,
			Meanings:    []string{"void", "chaos", "transformation"},
			VisualID:    "base_void_visual",
			RitualModifiers: map[string]float64{
				"power":     1.5,
				"stability": 0.5,
				"duration":  1.3,
			},
			WorldEffects: map[string]float64{
				"reality_distortion": 1.0,
			},
		},
	}

	// Save base symbols
	for _, symbol := range baseSymbols {
		filePath := filepath.Join(basePath, symbol.ID+".json")
		data, err := json.MarshalIndent(symbol, "", "  ")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDefaultPatterns creates default symbol patterns
func (sr *SymbolRegistry) CreateDefaultPatterns(patternsPath string) error {
	// Create base patterns
	patterns := []SymbolPattern{
		{
			ID:        "pattern_circle",
			BaseShape: "circle",
			Elements: []SymbolElement{
				{
					Type:     "circle",
					Position: [2]float64{0.5, 0.5},
					Size:     0.8,
					Rotation: 0,
					Color:    "primary",
				},
			},
			Transformations: []SymbolTransformation{
				{
					Type:   "none",
					Params: map[string]float64{},
				},
			},
			ColorScheme: []string{"#ff0000", "#00ff00", "#0000ff"},
			GenerationRules: map[string]interface{}{
				"complexity": 0.3,
				"symmetry":   "radial",
			},
		},
		{
			ID:        "pattern_triangle",
			BaseShape: "triangle",
			Elements: []SymbolElement{
				{
					Type:     "triangle",
					Position: [2]float64{0.5, 0.5},
					Size:     0.8,
					Rotation: 0,
					Color:    "primary",
				},
			},
			Transformations: []SymbolTransformation{
				{
					Type: "rotate",
					Params: map[string]float64{
						"angle": 0,
					},
				},
			},
			ColorScheme: []string{"#ffff00", "#ff00ff", "#00ffff"},
			GenerationRules: map[string]interface{}{
				"complexity": 0.4,
				"symmetry":   "reflective",
			},
		},
		{
			ID:        "pattern_rune",
			BaseShape: "line",
			Elements: []SymbolElement{
				{
					Type:     "line",
					Position: [2]float64{0.3, 0.5},
					Size:     0.6,
					Rotation: 90,
					Color:    "primary",
				},
				{
					Type:     "line",
					Position: [2]float64{0.5, 0.3},
					Size:     0.4,
					Rotation: 0,
					Color:    "secondary",
				},
			},
			Transformations: []SymbolTransformation{
				{
					Type:   "none",
					Params: map[string]float64{},
				},
			},
			ColorScheme: []string{"#ffffff", "#aaaaaa", "#444444"},
			GenerationRules: map[string]interface{}{
				"complexity": 0.6,
				"symmetry":   "none",
			},
		},
		{
			ID:        "pattern_star",
			BaseShape: "star",
			Elements: []SymbolElement{
				{
					Type:     "star",
					Position: [2]float64{0.5, 0.5},
					Size:     0.7,
					Rotation: 0,
					Color:    "primary",
					Params: map[string]float64{
						"points":       5,
						"inner_radius": 0.4,
					},
				},
			},
			Transformations: []SymbolTransformation{
				{
					Type: "rotate",
					Params: map[string]float64{
						"angle": 0,
					},
				},
			},
			ColorScheme: []string{"#ffcc00", "#ff6600", "#cc3300"},
			GenerationRules: map[string]interface{}{
				"complexity": 0.5,
				"symmetry":   "radial",
			},
		},
	}

	// Save patterns
	for _, pattern := range patterns {
		filePath := filepath.Join(patternsPath, pattern.ID+".json")
		data, err := json.MarshalIndent(pattern, "", "  ")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDefaultMeanings creates default meaning groups
func (sr *SymbolRegistry) CreateDefaultMeanings(meaningsPath string) error {
	// Create meaning groups
	meanings := map[string][]string{
		"elemental": {"fire", "water", "earth", "air", "lightning", "ice", "metal", "wood"},
		"arcane":    {"magic", "knowledge", "wisdom", "power", "secrets", "mysteries", "divination", "enchantment"},
		"primal":    {"life", "death", "growth", "decay", "birth", "age", "strength", "weakness"},
		"void":      {"chaos", "order", "creation", "destruction", "transformation", "void", "darkness", "light"},
		"emotions":  {"fear", "courage", "love", "hate", "joy", "sorrow", "anger", "peace"},
		"concepts":  {"time", "space", "reality", "dream", "mind", "body", "spirit", "matter"},
	}

	// Save meanings
	data, err := json.MarshalIndent(meanings, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(meaningsPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(meaningsPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// LoadState loads the state of the symbol registry
func (sr *SymbolRegistry) LoadState() error {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	symbolsPath := filepath.Join(sr.savePath, "symbols.json")

	// Check if file exists
	if _, err := os.Stat(symbolsPath); os.IsNotExist(err) {
		return fmt.Errorf("symbols file does not exist")
	}

	// Load file
	data, err := ioutil.ReadFile(symbolsPath)
	if err != nil {
		return err
	}

	var symbols []*Symbol
	err = json.Unmarshal(data, &symbols)
	if err != nil {
		return err
	}

	// Clear existing symbols
	sr.symbols = make(map[string]*Symbol)
	sr.discoveredSymbols = make(map[string]*Symbol)
	sr.symbolsByType = make(map[string][]*Symbol)

	// Add loaded symbols
	for _, symbol := range symbols {
		sr.symbols[symbol.ID] = symbol

		// Add to symbols by type
		if _, exists := sr.symbolsByType[symbol.SymbolType]; !exists {
			sr.symbolsByType[symbol.SymbolType] = make([]*Symbol, 0)
		}
		sr.symbolsByType[symbol.SymbolType] = append(sr.symbolsByType[symbol.SymbolType], symbol)

		// Add to discovered symbols if discovered
		if symbol.IsDiscovered {
			sr.discoveredSymbols[symbol.ID] = symbol
		}
	}

	return nil
}

// SaveState saves the state of the symbol registry
func (sr *SymbolRegistry) SaveState() error {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	symbolsPath := filepath.Join(sr.savePath, "symbols.json")

	// Make sure directory exists
	dir := filepath.Dir(symbolsPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Collect all symbols
	symbols := make([]*Symbol, 0, len(sr.symbols))
	for _, symbol := range sr.symbols {
		symbols = append(symbols, symbol)
	}

	// Serialize and save
	data, err := json.MarshalIndent(symbols, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(symbolsPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// AddSymbol adds a symbol to the registry
func (sr *SymbolRegistry) AddSymbol(symbol *Symbol) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	sr.symbols[symbol.ID] = symbol

	// Add to symbols by type
	if _, exists := sr.symbolsByType[symbol.SymbolType]; !exists {
		sr.symbolsByType[symbol.SymbolType] = make([]*Symbol, 0)
	}
	sr.symbolsByType[symbol.SymbolType] = append(sr.symbolsByType[symbol.SymbolType], symbol)

	// Add to discovered symbols if discovered
	if symbol.IsDiscovered {
		sr.discoveredSymbols[symbol.ID] = symbol
	}
}

// GetSymbol returns a symbol by ID
func (sr *SymbolRegistry) GetSymbol(id string) *Symbol {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	return sr.symbols[id]
}

// GetAllSymbols returns all symbols
func (sr *SymbolRegistry) GetAllSymbols() []*Symbol {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	symbols := make([]*Symbol, 0, len(sr.symbols))
	for _, symbol := range sr.symbols {
		symbols = append(symbols, symbol)
	}

	return symbols
}

// GetDiscoveredSymbols returns all discovered symbols
func (sr *SymbolRegistry) GetDiscoveredSymbols() []*Symbol {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	symbols := make([]*Symbol, 0, len(sr.discoveredSymbols))
	for _, symbol := range sr.discoveredSymbols {
		symbols = append(symbols, symbol)
	}

	return symbols
}

// GetSymbolsByType returns symbols of a specific type
func (sr *SymbolRegistry) GetSymbolsByType(symbolType string) []*Symbol {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	symbols, exists := sr.symbolsByType[symbolType]
	if !exists {
		return make([]*Symbol, 0)
	}

	return symbols
}

// Helper functions for ritual registry

// LoadBaseRituals loads base rituals from files
func (rr *RitualRegistry) LoadBaseRituals() error {
	basePath := filepath.Join(rr.savePath, "base")

	// Create directory if it doesn't exist
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err = os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			return err
		}

		// Create default base rituals
		err = rr.CreateDefaultBaseRituals(basePath)
		if err != nil {
			return err
		}
	}

	// Load files
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".json" {
			// Load ritual
			filePath := filepath.Join(basePath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			var ritual Ritual
			err = json.Unmarshal(data, &ritual)
			if err != nil {
				continue
			}

			// Add to base rituals
			rr.baseRituals = append(rr.baseRituals, &ritual)
		}
	}

	// Load effect templates
	effectsPath := filepath.Join(rr.savePath, "effects")
	if _, err := os.Stat(effectsPath); os.IsNotExist(err) {
		err = os.MkdirAll(effectsPath, os.ModePerm)
		if err != nil {
			return err
		}

		// Create default effects
		err = rr.CreateDefaultEffects(effectsPath)
		if err != nil {
			return err
		}
	}

	// Load effect files
	files, err = ioutil.ReadDir(effectsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".json" {
			// Load effect
			filePath := filepath.Join(effectsPath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}

			var effect RitualEffect
			err = json.Unmarshal(data, &effect)
			if err != nil {
				continue
			}

			// Add to effect templates
			rr.effectTemplates[file.Name()[:len(file.Name())-5]] = effect
		}
	}

	return nil
}

// CreateDefaultBaseRituals creates default base rituals
func (rr *RitualRegistry) CreateDefaultBaseRituals(basePath string) error {
	// Create base rituals for different locations
	baseRituals := []Ritual{
		{
			ID:               "base_forest",
			Name:             "Forest Ritual",
			Description:      "A ritual performed in the forest",
			RequiredSymbols:  []string{},
			RequiredItems:    []string{},
			RequiredLocation: "forest",
			Actions:          []string{},
			Difficulty:       0.5,
			TimeRequired:     60,
			Effects:          []RitualEffect{},
			SuccessChance:    0.7,
			FailureEffects:   []RitualEffect{},
			EvolutionPath:    []string{},
		},
		{
			ID:               "base_water",
			Name:             "Water Ritual",
			Description:      "A ritual performed near water",
			RequiredSymbols:  []string{},
			RequiredItems:    []string{},
			RequiredLocation: "water",
			Actions:          []string{},
			Difficulty:       0.6,
			TimeRequired:     90,
			Effects:          []RitualEffect{},
			SuccessChance:    0.65,
			FailureEffects:   []RitualEffect{},
			EvolutionPath:    []string{},
		},
		{
			ID:               "base_cave",
			Name:             "Cave Ritual",
			Description:      "A ritual performed in a cave",
			RequiredSymbols:  []string{},
			RequiredItems:    []string{},
			RequiredLocation: "cave",
			Actions:          []string{},
			Difficulty:       0.7,
			TimeRequired:     120,
			Effects:          []RitualEffect{},
			SuccessChance:    0.6,
			FailureEffects:   []RitualEffect{},
			EvolutionPath:    []string{},
		},
		{
			ID:               "base_clearing",
			Name:             "Clearing Ritual",
			Description:      "A ritual performed in an open clearing",
			RequiredSymbols:  []string{},
			RequiredItems:    []string{},
			RequiredLocation: "clearing",
			Actions:          []string{},
			Difficulty:       0.4,
			TimeRequired:     45,
			Effects:          []RitualEffect{},
			SuccessChance:    0.75,
			FailureEffects:   []RitualEffect{},
			EvolutionPath:    []string{},
		},
		{
			ID:               "base_hill",
			Name:             "Hilltop Ritual",
			Description:      "A ritual performed on a hill",
			RequiredSymbols:  []string{},
			RequiredItems:    []string{},
			RequiredLocation: "hill",
			Actions:          []string{},
			Difficulty:       0.5,
			TimeRequired:     60,
			Effects:          []RitualEffect{},
			SuccessChance:    0.7,
			FailureEffects:   []RitualEffect{},
			EvolutionPath:    []string{},
		},
	}

	// Save base rituals
	for _, ritual := range baseRituals {
		filePath := filepath.Join(basePath, ritual.ID+".json")
		data, err := json.MarshalIndent(ritual, "", "  ")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDefaultEffects creates default effect templates
func (rr *RitualRegistry) CreateDefaultEffects(effectsPath string) error {
	// Create effect templates
	effects := []struct {
		ID     string
		Effect RitualEffect
	}{
		{
			ID: "metamorphosis_minor",
			Effect: RitualEffect{
				Type:           "metamorphosis",
				Target:         "area",
				Value:          0.3,
				Duration:       1800, // 30 minutes
				Tags:           []string{"metamorphosis", "minor"},
				Description:    "Creates a minor metamorphosis in the surrounding area",
				MetamorphOrder: 1,
				MetamorphArea:  10.0,
			},
		},
		{
			ID: "metamorphosis_moderate",
			Effect: RitualEffect{
				Type:           "metamorphosis",
				Target:         "area",
				Value:          0.6,
				Duration:       3600, // 1 hour
				Tags:           []string{"metamorphosis", "moderate"},
				Description:    "Creates a moderate metamorphosis in the surrounding area",
				MetamorphOrder: 2,
				MetamorphArea:  20.0,
			},
		},
		{
			ID: "metamorphosis_major",
			Effect: RitualEffect{
				Type:           "metamorphosis",
				Target:         "area",
				Value:          0.9,
				Duration:       7200, // 2 hours
				Tags:           []string{"metamorphosis", "major"},
				Description:    "Creates a major metamorphosis in the surrounding area",
				MetamorphOrder: 3,
				MetamorphArea:  30.0,
			},
		},
		{
			ID: "player_heal",
			Effect: RitualEffect{
				Type:        "player",
				Target:      "health",
				Value:       50.0,
				Duration:    0, // Instant
				Tags:        []string{"healing", "positive"},
				Description: "Heals the player",
			},
		},
		{
			ID: "player_sanity",
			Effect: RitualEffect{
				Type:        "player",
				Target:      "sanity",
				Value:       30.0,
				Duration:    0, // Instant
				Tags:        []string{"mental", "positive"},
				Description: "Restores some of the player's sanity",
			},
		},
		{
			ID: "spawn_friendly",
			Effect: RitualEffect{
				Type:            "spawn",
				Target:          "entity",
				Value:           1.0,
				Duration:        1800, // 30 minutes
				Tags:            []string{"entity", "friendly"},
				Description:     "Summons a friendly entity",
				SpawnEntityType: "friendly_spirit",
				SpawnCount:      1,
				SpawnRadius:     5.0,
			},
		},
		{
			ID: "spawn_hostile",
			Effect: RitualEffect{
				Type:            "spawn",
				Target:          "entity",
				Value:           1.0,
				Duration:        1800, // 30 minutes
				Tags:            []string{"entity", "hostile"},
				Description:     "Summons a hostile entity",
				SpawnEntityType: "hostile_spirit",
				SpawnCount:      1,
				SpawnRadius:     10.0,
			},
		},
		{
			ID: "item_create",
			Effect: RitualEffect{
				Type:        "item",
				Target:      "create",
				Value:       1.0,
				Duration:    0, // Instant
				Tags:        []string{"item", "create"},
				Description: "Creates a mystical item",
				ItemID:      "ritual_item",
			},
		},
		{
			ID: "weather_change",
			Effect: RitualEffect{
				Type:        "weather",
				Target:      "local",
				Value:       1.0,
				Duration:    3600, // 1 hour
				Tags:        []string{"weather", "environment"},
				Description: "Changes the local weather",
			},
		},
		{
			ID: "reveal_symbols",
			Effect: RitualEffect{
				Type:        "knowledge",
				Target:      "symbols",
				Value:       0.2,
				Duration:    0, // Instant
				Tags:        []string{"knowledge", "symbols"},
				Description: "Reveals information about nearby symbols",
			},
		},
	}

	// Save effects
	for _, effectData := range effects {
		filePath := filepath.Join(effectsPath, effectData.ID+".json")
		data, err := json.MarshalIndent(effectData.Effect, "", "  ")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadState loads the state of the ritual registry
func (rr *RitualRegistry) LoadState() error {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	ritualsPath := filepath.Join(rr.savePath, "rituals.json")

	// Check if file exists
	if _, err := os.Stat(ritualsPath); os.IsNotExist(err) {
		return fmt.Errorf("rituals file does not exist")
	}

	// Load file
	data, err := ioutil.ReadFile(ritualsPath)
	if err != nil {
		return err
	}

	var rituals []*Ritual
	err = json.Unmarshal(data, &rituals)
	if err != nil {
		return err
	}

	// Clear existing rituals
	rr.rituals = make(map[string]*Ritual)
	rr.discoveredRituals = make(map[string]*Ritual)
	rr.ritualsBySymbol = make(map[string][]*Ritual)
	rr.ritualsByEffect = make(map[string][]*Ritual)

	// Add loaded rituals
	for _, ritual := range rituals {
		rr.rituals[ritual.ID] = ritual

		// Add to rituals by symbol
		for _, symbolID := range ritual.RequiredSymbols {
			if _, exists := rr.ritualsBySymbol[symbolID]; !exists {
				rr.ritualsBySymbol[symbolID] = make([]*Ritual, 0)
			}
			rr.ritualsBySymbol[symbolID] = append(rr.ritualsBySymbol[symbolID], ritual)
		}

		// Add to rituals by effect type
		for _, effect := range ritual.Effects {
			if _, exists := rr.ritualsByEffect[effect.Type]; !exists {
				rr.ritualsByEffect[effect.Type] = make([]*Ritual, 0)
			}
			rr.ritualsByEffect[effect.Type] = append(rr.ritualsByEffect[effect.Type], ritual)
		}

		// Add to discovered rituals if discovered
		if ritual.IsDiscovered {
			rr.discoveredRituals[ritual.ID] = ritual
		}
	}

	return nil
}

// SaveState saves the state of the ritual registry
func (rr *RitualRegistry) SaveState() error {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	ritualsPath := filepath.Join(rr.savePath, "rituals.json")

	// Make sure directory exists
	dir := filepath.Dir(ritualsPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Collect all rituals
	rituals := make([]*Ritual, 0, len(rr.rituals))
	for _, ritual := range rr.rituals {
		rituals = append(rituals, ritual)
	}

	// Serialize and save
	data, err := json.MarshalIndent(rituals, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(ritualsPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// AddRitual adds a ritual to the registry
func (rr *RitualRegistry) AddRitual(ritual *Ritual) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	rr.rituals[ritual.ID] = ritual

	// Add to rituals by symbol
	for _, symbolID := range ritual.RequiredSymbols {
		if _, exists := rr.ritualsBySymbol[symbolID]; !exists {
			rr.ritualsBySymbol[symbolID] = make([]*Ritual, 0)
		}
		rr.ritualsBySymbol[symbolID] = append(rr.ritualsBySymbol[symbolID], ritual)
	}

	// Add to rituals by effect type
	for _, effect := range ritual.Effects {
		if _, exists := rr.ritualsByEffect[effect.Type]; !exists {
			rr.ritualsByEffect[effect.Type] = make([]*Ritual, 0)
		}
		rr.ritualsByEffect[effect.Type] = append(rr.ritualsByEffect[effect.Type], ritual)
	}

	// Add to discovered rituals if discovered
	if ritual.IsDiscovered {
		rr.discoveredRituals[ritual.ID] = ritual
	}
}

// GetRitual returns a ritual by ID
func (rr *RitualRegistry) GetRitual(id string) *Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	return rr.rituals[id]
}

// GetAllRituals returns all rituals
func (rr *RitualRegistry) GetAllRituals() []*Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	rituals := make([]*Ritual, 0, len(rr.rituals))
	for _, ritual := range rr.rituals {
		rituals = append(rituals, ritual)
	}

	return rituals
}

// GetDiscoveredRituals returns all discovered rituals
func (rr *RitualRegistry) GetDiscoveredRituals() []*Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	rituals := make([]*Ritual, 0, len(rr.discoveredRituals))
	for _, ritual := range rr.discoveredRituals {
		rituals = append(rituals, ritual)
	}

	return rituals
}

// GetUndiscoveredRituals returns all undiscovered rituals
func (rr *RitualRegistry) GetUndiscoveredRituals() []*Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	rituals := make([]*Ritual, 0)
	for _, ritual := range rr.rituals {
		if !ritual.IsDiscovered {
			rituals = append(rituals, ritual)
		}
	}

	return rituals
}

// GetRitualsBySymbol returns rituals that require a specific symbol
func (rr *RitualRegistry) GetRitualsBySymbol(symbolID string) []*Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	rituals, exists := rr.ritualsBySymbol[symbolID]
	if !exists {
		return make([]*Ritual, 0)
	}

	return rituals
}

// GetRitualsByEffect returns rituals that produce a specific effect type
func (rr *RitualRegistry) GetRitualsByEffect(effectType string) []*Ritual {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	rituals, exists := rr.ritualsByEffect[effectType]
	if !exists {
		return make([]*Ritual, 0)
	}

	return rituals
}

// Utility functions

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result[i] = chars[r.Intn(len(chars))]
	}

	return string(result)
}

// generateSymbolNameSuffix generates a suffix for a symbol name based on type
func generateSymbolNameSuffix(symbolType string, r *rand.Rand) string {
	switch symbolType {
	case "elemental":
		suffixes := []string{"the Elements", "the Four Winds", "the Earth", "the Flame", "the Tides", "the Storm"}
		return suffixes[r.Intn(len(suffixes))]
	case "arcane":
		suffixes := []string{"the Arcane", "the Hidden Knowledge", "the Forbidden", "the Ancient Ones", "the Stars"}
		return suffixes[r.Intn(len(suffixes))]
	case "primal":
		suffixes := []string{"the Wild", "Life and Death", "the Beast", "the Seasons", "Growth", "the Forest"}
		return suffixes[r.Intn(len(suffixes))]
	case "void":
		suffixes := []string{"the Void", "the Abyss", "Chaos", "the Outer Dark", "the Between", "Nothingness"}
		return suffixes[r.Intn(len(suffixes))]
	default:
		suffixes := []string{"Mystery", "the Unknown", "Power", "Wisdom", "Transformation", "Secrets"}
		return suffixes[r.Intn(len(suffixes))]
	}
}

// generateSymbolAdjective generates an adjective for a symbol based on type
func generateSymbolAdjective(symbolType string, r *rand.Rand) string {
	switch symbolType {
	case "elemental":
		adjectives := []string{"shimmering", "vibrant", "pulsing", "resonant", "natural", "primal"}
		return adjectives[r.Intn(len(adjectives))]
	case "arcane":
		adjectives := []string{"complex", "intricate", "glowing", "mysterious", "cryptic", "enigmatic"}
		return adjectives[r.Intn(len(adjectives))]
	case "primal":
		adjectives := []string{"organic", "living", "wild", "untamed", "natural", "primitive"}
		return adjectives[r.Intn(len(adjectives))]
	case "void":
		adjectives := []string{"unsettling", "dark", "shadowy", "distorted", "otherworldly", "alien"}
		return adjectives[r.Intn(len(adjectives))]
	default:
		adjectives := []string{"strange", "unusual", "powerful", "ancient", "forgotten", "hidden"}
		return adjectives[r.Intn(len(adjectives))]
	}
}

// generateSymbolVerb generates a verb phrase for a symbol based on type
func generateSymbolVerb(symbolType string, r *rand.Rand) string {
	switch symbolType {
	case "elemental":
		verbs := []string{
			"shifts like flowing water",
			"flickers like flame",
			"seems to move with the wind",
			"pulses with earthen power",
			"crackles with subtle energy",
		}
		return verbs[r.Intn(len(verbs))]
	case "arcane":
		verbs := []string{
			"contains patterns within patterns",
			"seems to change when not directly observed",
			"emits a subtle glow in darkness",
			"whispers ancient knowledge",
			"pulls at the mind like a puzzle",
		}
		return verbs[r.Intn(len(verbs))]
	case "primal":
		verbs := []string{
			"grows subtly like living tissue",
			"pulses with the rhythm of a heartbeat",
			"changes with the seasons",
			"calls to the wild parts of the mind",
			"feels warm to the touch like a living thing",
		}
		return verbs[r.Intn(len(verbs))]
	case "void":
		verbs := []string{
			"seems to absorb light around it",
			"distorts perception when viewed directly",
			"fills the mind with strange thoughts",
			"shifts in ways that shouldn't be possible",
			"feels wrong in a way that's hard to describe",
		}
		return verbs[r.Intn(len(verbs))]
	default:
		verbs := []string{
			"draws the eye in unusual ways",
			"feels significant somehow",
			"seems older than its surroundings",
			"stands out despite its simplicity",
			"resonates with unseen power",
		}
		return verbs[r.Intn(len(verbs))]
	}
}

// generateSymbolEffect generates an effect description for a symbol based on type
func generateSymbolEffect(symbolType string, r *rand.Rand) string {
	switch symbolType {
	case "elemental":
		effects := []string{
			"alter the flow of natural energies around it",
			"harmonize with elemental forces",
			"change subtly with the weather",
			"amplify natural phenomena nearby",
			"balance opposing elemental forces",
		}
		return effects[r.Intn(len(effects))]
	case "arcane":
		effects := []string{
			"contain knowledge from bygone ages",
			"unlock hidden potential in the mind",
			"preserve magical energies for later use",
			"protect against malevolent forces",
			"enhance ritual workings significantly",
		}
		return effects[r.Intn(len(effects))]
	case "primal":
		effects := []string{
			"connect deeply with the cycles of life and death",
			"accelerate growth in nearby plants",
			"calm or agitate animals nearby",
			"draw strength from the natural world",
			"enhance vitality in living beings",
		}
		return effects[r.Intn(len(effects))]
	case "void":
		effects := []string{
			"thin the boundaries between realities",
			"distort space and perception around it",
			"whisper thoughts from beyond normal reality",
			"create feelings of unease and dread",
			"transform matter in unpredictable ways",
		}
		return effects[r.Intn(len(effects))]
	default:
		effects := []string{
			"hold power waiting to be unlocked",
			"connect to something beyond understanding",
			"resonate with mysterious forces",
			"contain knowledge lost to time",
			"transform those who study it deeply",
		}
		return effects[r.Intn(len(effects))]
	}
}

// generateSymbolMeanings generates meaning for a symbol based on type
func generateSymbolMeanings(symbolType string, r *rand.Rand, count int) []string {
	// Define meanings by type
	meaningsByType := map[string][]string{
		"elemental": {"fire", "water", "earth", "air", "lightning", "ice", "metal", "wood", "crystal", "magma", "smoke", "steam"},
		"arcane":    {"magic", "knowledge", "wisdom", "power", "secrets", "mysteries", "divination", "enchantment", "illusion", "transformation", "binding", "summoning"},
		"primal":    {"life", "death", "growth", "decay", "birth", "age", "strength", "weakness", "predator", "prey", "fertility", "famine"},
		"void":      {"chaos", "order", "creation", "destruction", "void", "darkness", "light", "beginning", "end", "beyond", "between", "outside"},
		"default":   {"mystery", "unknown", "power", "wisdom", "change", "stasis", "harmony", "discord", "balance", "excess", "scarcity", "abundance"},
	}

	// Get meanings for this type
	meanings, exists := meaningsByType[symbolType]
	if !exists {
		meanings = meaningsByType["default"]
	}

	// Add some crossover from other categories
	allMeanings := make([]string, 0)
	for _, typeMeanings := range meaningsByType {
		allMeanings = append(allMeanings, typeMeanings...)
	}

	// Create a shuffled copy of meanings
	shuffled := make([]string, len(meanings))
	copy(shuffled, meanings)

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Select count meanings, with 70% chance from primary type and 30% from others
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if i < len(shuffled) && r.Float64() < 0.7 {
			// Select from primary type
			result = append(result, shuffled[i])
		} else {
			// Select from all types
			idx := r.Intn(len(allMeanings))
			result = append(result, allMeanings[idx])
		}
	}

	// Remove duplicates
	return uniqueStrings(result)
}

// generateRitualNameSuffix generates a suffix for a ritual name
func generateRitualNameSuffix(symbol *Symbol, r *rand.Rand) string {
	// If we have a symbol, base name on its meanings
	if symbol != nil && len(symbol.Meanings) > 0 {
		// Capitalize first letter
		meaning := symbol.Meanings[r.Intn(len(symbol.Meanings))]
		if len(meaning) > 0 {
			meaning = strings.Title(meaning)
		}
		return meaning
	}

	// Otherwise, use generic names
	suffixes := []string{
		"the Ancient Ones", "the Hidden Truth", "the Veil", "the Path",
		"Awakening", "Binding", "Communion", "Revelation",
		"the Forgotten", "the Unseen", "the Beyond", "the Depths",
	}
	return suffixes[r.Intn(len(suffixes))]
}

// generateRitualAdjective generates an adjective for a ritual based on location
func generateRitualAdjective(location string, r *rand.Rand) string {
	switch location {
	case "forest":
		adjectives := []string{"verdant", "sylvan", "ancient", "whispering", "shadowed", "wild"}
		return adjectives[r.Intn(len(adjectives))]
	case "water":
		adjectives := []string{"flowing", "reflective", "tidal", "deep", "purifying", "transformative"}
		return adjectives[r.Intn(len(adjectives))]
	case "cave":
		adjectives := []string{"echoing", "darkened", "hidden", "primordial", "resonant", "secret"}
		return adjectives[r.Intn(len(adjectives))]
	case "clearing":
		adjectives := []string{"open", "starlit", "exposed", "circular", "sacred", "balanced"}
		return adjectives[r.Intn(len(adjectives))]
	case "hill":
		adjectives := []string{"elevated", "windswept", "overlooking", "ancient", "watching", "boundary"}
		return adjectives[r.Intn(len(adjectives))]
	default:
		adjectives := []string{"mysterious", "powerful", "arcane", "secret", "forgotten", "complex"}
		return adjectives[r.Intn(len(adjectives))]
	}
}

// generateRitualEffect generates an effect description for a ritual based on location
func generateRitualEffect(location string, r *rand.Rand) string {
	switch location {
	case "forest":
		effects := []string{
			"channels the ancient power of the trees",
			"awakens the spirits of the forest",
			"binds the performer to the cycles of growth",
			"parts the veil between our world and the green realm",
			"reveals the hidden paths through the wilderness",
		}
		return effects[r.Intn(len(effects))]
	case "water":
		effects := []string{
			"calls upon the transformative power of water",
			"cleanses the spirit of impurities",
			"reveals hidden depths of knowledge",
			"flows between worlds like the tide",
			"reflects truths otherwise hidden from sight",
		}
		return effects[r.Intn(len(effects))]
	case "cave":
		effects := []string{
			"echoes with the whispers of the earth",
			"reveals what lies beneath the surface of reality",
			"unlocks ancient secrets entombed in stone",
			"opens passages to the underworld",
			"resonates with the beating heart of the world",
			"opens passages to the underworld",
			"resonates with the beating heart of the world",
		}
		return effects[r.Intn(len(effects))]
	case "clearing":
		effects := []string{
			"draws power from the open sky above",
			"creates a perfect circle of energy",
			"balances opposing forces in harmony",
			"opens the performer to cosmic influences",
			"establishes a neutral ground between worlds",
		}
		return effects[r.Intn(len(effects))]
	case "hill":
		effects := []string{
			"harnesses the power of the liminal space between earth and sky",
			"surveys the landscape of both physical and spiritual realms",
			"draws energy from the heights and depths simultaneously",
			"establishes dominion over the surrounding territory",
			"sends messages to distant powers",
		}
		return effects[r.Intn(len(effects))]
	default:
		effects := []string{
			"channels powers beyond mortal understanding",
			"transforms the performer in unexpected ways",
			"reveals hidden truths about the world",
			"creates a bridge between different realms",
			"bends reality according to the performer's will",
		}
		return effects[r.Intn(len(effects))]
	}
}

// generateLocationDescription generates a description for a ritual location
func generateLocationDescription(location string) string {
	switch location {
	case "forest":
		return "a dense area of trees with dappled sunlight"
	case "water":
		return "a body of water like a stream, lake, or pond"
	case "cave":
		return "a dark underground space away from the sun"
	case "clearing":
		return "an open area free from trees and obstructions"
	case "hill":
		return "an elevated position with a view of the surroundings"
	default:
		return "a special place with the right energies"
	}
}

// generateRitualItem generates a ritual item description
func generateRitualItem(itemType string, r *rand.Rand) string {
	switch itemType {
	case "herb":
		herbs := []string{
			"wild sage", "moonflower", "bloodroot", "ghost moss", "twisted bramble",
			"whispering fern", "black lotus", "dreamer's weed", "starleaf", "nightshade",
		}
		return herbs[r.Intn(len(herbs))]
	case "mineral":
		minerals := []string{
			"quartz crystal", "red ochre", "black salt", "fool's gold", "lodestone",
			"obsidian shard", "amber fragment", "cave pearl", "thunderstone", "silver dust",
		}
		return minerals[r.Intn(len(minerals))]
	case "bone":
		bones := []string{
			"small animal skull", "bird bone", "vertebrae", "antler fragment", "tooth",
			"jawbone", "carved bone", "hollow bone", "charred bone", "ancient remains",
		}
		return bones[r.Intn(len(bones))]
	case "fluid":
		fluids := []string{
			"clear spring water", "morning dew", "blood (animal)", "rendered fat", "tree sap",
			"fermented berries", "pine resin", "mushroom extract", "flower essence", "rainwater",
		}
		return fluids[r.Intn(len(fluids))]
	case "cloth":
		cloths := []string{
			"red cloth strip", "black silk square", "woven grass mat", "dyed linen", "burial shroud",
			"embroidered patch", "unspun wool", "spider silk", "ceremonial banner", "charred rags",
		}
		return cloths[r.Intn(len(cloths))]
	case "tool":
		tools := []string{
			"bone needle", "stone knife", "wooden bowl", "copper wire", "clay vessel",
			"leather pouch", "glass vial", "carved stick", "stone mortar", "wooden flute",
		}
		return tools[r.Intn(len(tools))]
	default:
		items := []string{
			"mysterious artifact", "symbolic object", "personal token", "natural rarity", "found curiosity",
		}
		return items[r.Intn(len(items))]
	}
}

// generateRitualAction generates a ritual action description
func generateRitualAction(actionType string, r *rand.Rand) string {
	switch actionType {
	case "draw":
		draws := []string{
			"Draw a circle on the ground",
			"Trace the symbol in the air",
			"Mark your forehead with the symbol",
			"Etch the rune into nearby wood",
			"Draw connecting lines between ritual items",
		}
		return draws[r.Intn(len(draws))]
	case "place":
		places := []string{
			"Place items at the cardinal points",
			"Arrange objects in the shape of the symbol",
			"Create a small altar with the items",
			"Position offerings in a specific pattern",
			"Set up a boundary of ritual components",
		}
		return places[r.Intn(len(places))]
	case "chant":
		chants := []string{
			"Recite the names of the symbols three times",
			"Chant a rhythmic invocation",
			"Whisper secret words into cupped hands",
			"Sing a melody that comes to mind",
			"Speak the ritual phrases backward",
		}
		return chants[r.Intn(len(chants))]
	case "burn":
		burns := []string{
			"Burn a small offering in the center",
			"Light candles at specific points",
			"Set fire to herbs to create sacred smoke",
			"Burn a drawing of the symbol",
			"Create a fire in the ritual space",
		}
		return burns[r.Intn(len(burns))]
	case "pour":
		pours := []string{
			"Pour liquid in a circle around you",
			"Create a line of fluid between components",
			"Anoint each ritual item with drops",
			"Form the symbol using poured liquid",
			"Sprinkle water in the four directions",
		}
		return pours[r.Intn(len(pours))]
	case "scatter":
		scatters := []string{
			"Scatter herbs or dust in a pattern",
			"Cast small objects from hand to ground",
			"Distribute offerings around the space",
			"Throw ashes into the air",
			"Spread materials along a prepared path",
		}
		return scatters[r.Intn(len(scatters))]
	case "meditate":
		meditations := []string{
			"Close your eyes and visualize the symbols",
			"Enter a trance state for several minutes",
			"Focus on your breath until time seems to slow",
			"Empty your mind and await visions",
			"Concentrate on the purpose of the ritual",
		}
		return meditations[r.Intn(len(meditations))]
	case "gesture":
		gestures := []string{
			"Make specific hand signs toward each direction",
			"Move in a circular pattern around the ritual space",
			"Perform a series of symbolic gestures",
			"Mimic the movements of certain animals",
			"Wave hands over ritual components in sequence",
		}
		return gestures[r.Intn(len(gestures))]
	default:
		actions := []string{
			"Perform the actions that feel most appropriate",
			"Follow your intuition in the moment",
			"Complete the ritual as the symbols dictate",
			"Do what seems necessary to complete the working",
			"Let the ritual guide your movements",
		}
		return actions[r.Intn(len(actions))]
	}
}

// generatePrimaryEffect generates a primary effect for a ritual based on location
func generatePrimaryEffect(location string, power float64, r *rand.Rand) RitualEffect {
	// Determine effect type based on location
	var effectType string
	var target string
	var description string

	switch location {
	case "forest":
		effectTypes := []string{"metamorphosis", "spawn", "player"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "metamorphosis" {
			target = "area"
			description = "Creates a nature-based metamorphosis in the surrounding forest"
		} else if effectType == "spawn" {
			target = "entity"
			description = "Summons a forest entity to aid the performer"
		} else {
			target = "health"
			description = "Channels the forest's vitality to heal"
		}

	case "water":
		effectTypes := []string{"metamorphosis", "player", "item"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "metamorphosis" {
			target = "area"
			description = "Creates a fluid, transformative metamorphosis around the water"
		} else if effectType == "player" {
			target = "sanity"
			description = "Cleanses the mind and restores sanity"
		} else {
			target = "create"
			description = "Manifests a mystical item from the water's essence"
		}

	case "cave":
		effectTypes := []string{"metamorphosis", "knowledge", "spawn"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "metamorphosis" {
			target = "area"
			description = "Creates a deep, earth-based metamorphosis in the cave"
		} else if effectType == "knowledge" {
			target = "symbols"
			description = "Reveals hidden knowledge from the depths"
		} else {
			target = "entity"
			description = "Calls forth entities from beneath the earth"
		}

	case "clearing":
		effectTypes := []string{"weather", "player", "metamorphosis"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "weather" {
			target = "local"
			description = "Changes the weather patterns above the clearing"
		} else if effectType == "player" {
			targets := []string{"health", "sanity"}
			target = targets[r.Intn(len(targets))]
			description = "Draws cosmic energy to restore the performer"
		} else {
			target = "area"
			description = "Creates an open, sky-based metamorphosis in the clearing"
		}

	case "hill":
		effectTypes := []string{"metamorphosis", "knowledge", "weather"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "metamorphosis" {
			target = "area"
			description = "Creates an elevated, far-reaching metamorphosis around the hill"
		} else if effectType == "knowledge" {
			target = "future"
			description = "Grants visions from the high vantage point"
		} else {
			target = "regional"
			description = "Influences weather patterns across the region"
		}

	default:
		effectTypes := []string{"metamorphosis", "player", "spawn"}
		effectType = effectTypes[r.Intn(len(effectTypes))]

		if effectType == "metamorphosis" {
			target = "area"
			description = "Creates a mysterious metamorphosis in the surrounding area"
		} else if effectType == "player" {
			targets := []string{"health", "sanity"}
			target = targets[r.Intn(len(targets))]
			description = "Channels unknown energies to empower the performer"
		} else {
			target = "entity"
			description = "Summons entities from beyond normal perception"
		}
	}

	// Create the effect with appropriate values based on power
	effect := RitualEffect{
		Type:        effectType,
		Target:      target,
		Tags:        []string{effectType, location},
		Description: description,
	}

	// Set specific values based on effect type
	switch effectType {
	case "metamorphosis":
		// Higher power means higher order metamorphosis
		order := 1
		if power > 0.4 {
			order = 2
		}
		if power > 0.7 {
			order = 3
		}

		// Duration and area depend on power
		duration := 1800.0 + power*3600.0 // 30 minutes to 1.5 hours
		area := 10.0 + power*30.0         // 10 to 40 meter radius

		effect.Value = 0.3 + power*0.7 // 0.3 to 1.0 intensity
		effect.Duration = duration
		effect.MetamorphOrder = order
		effect.MetamorphArea = area

	case "spawn":
		// Higher power means more or stronger entities
		entityType := "neutral_spirit"
		if power > 0.6 {
			entityType = "friendly_spirit"
		}

		spawnCount := 1 + int(power*3) // 1 to 4 entities

		effect.Value = 0.5 + power*0.5      // 0.5 to 1.0 strength
		effect.Duration = 1800 + power*3600 // 30 minutes to 1.5 hours
		effect.SpawnEntityType = entityType
		effect.SpawnCount = spawnCount
		effect.SpawnRadius = 5.0 + power*10.0

	case "player":
		// Higher power means more healing or restoration
		effect.Value = 30.0 + power*70.0 // 30 to 100 points
		effect.Duration = 0              // Instant effect

	case "item":
		effect.Value = 1.0
		effect.Duration = 0 // Instant effect

		// Item quality depends on power
		if power > 0.7 {
			effect.ItemID = "powerful_ritual_item"
		} else if power > 0.4 {
			effect.ItemID = "medium_ritual_item"
		} else {
			effect.ItemID = "minor_ritual_item"
		}

	case "knowledge":
		effect.Value = 0.2 + power*0.4 // 0.2 to 0.6 knowledge gain
		effect.Duration = 0            // Instant effect

	case "weather":
		effect.Value = 0.5 + power*0.5      // 0.5 to 1.0 intensity
		effect.Duration = 3600 + power*7200 // 1 to 3 hours
	}

	return effect
}

// generateSecondaryEffect generates a secondary effect for a ritual
func generateSecondaryEffect(location string, power float64, r *rand.Rand) RitualEffect {
	// Secondary effects are usually different from primary ones
	// For example, if primary is metamorphosis, secondary might be player effect

	effectTypes := []string{"player", "knowledge", "item", "weather"}
	effectType := effectTypes[r.Intn(len(effectTypes))]

	// Create the effect
	effect := RitualEffect{
		Type: effectType,
		Tags: []string{effectType, "secondary", location},
	}

	// Set specific values based on effect type
	switch effectType {
	case "player":
		targets := []string{"health", "sanity", "energy"}
		effect.Target = targets[r.Intn(len(targets))]
		effect.Description = fmt.Sprintf("A secondary effect that affects the performer's %s", effect.Target)
		effect.Value = 20.0 + power*40.0 // 20 to 60 points
		effect.Duration = 0              // Instant effect

	case "knowledge":
		targets := []string{"symbols", "rituals", "area"}
		effect.Target = targets[r.Intn(len(targets))]
		effect.Description = "Reveals additional knowledge as a side effect"
		effect.Value = 0.1 + power*0.3 // 0.1 to 0.4 knowledge gain
		effect.Duration = 0            // Instant effect

	case "item":
		effect.Target = "create"
		effect.Description = "Creates a minor item as a byproduct of the ritual"
		effect.Value = 1.0
		effect.Duration = 0 // Instant effect
		effect.ItemID = "minor_ritual_item"

	case "weather":
		effect.Target = "local"
		effect.Description = "Causes a subtle shift in local conditions"
		effect.Value = 0.3 + power*0.3      // 0.3 to 0.6 intensity
		effect.Duration = 1800 + power*3600 // 30 minutes to 1.5 hours
	}

	return effect
}

// generateFailureEffect generates an effect for a failed ritual
func generateFailureEffect(location string, power float64, r *rand.Rand) RitualEffect {
	// Failure effects are usually negative
	effectTypes := []string{"player_harm", "metamorphosis_unstable", "spawn_hostile"}
	effectType := effectTypes[r.Intn(len(effectTypes))]

	// Create the effect
	effect := RitualEffect{
		Type: effectType,
		Tags: []string{"failure", "negative", location},
	}

	// Set specific values based on effect type
	switch effectType {
	case "player_harm":
		targets := []string{"health", "sanity"}
		effect.Target = targets[r.Intn(len(targets))]
		effect.Description = fmt.Sprintf("The ritual backfires, damaging the performer's %s", effect.Target)
		effect.Value = -(10.0 + power*30.0) // -10 to -40 points (negative for damage)
		effect.Duration = 0                 // Instant effect

	case "metamorphosis_unstable":
		effect.Target = "area"
		effect.Description = "The ritual creates an unstable, chaotic metamorphosis"
		effect.Value = 0.2 + power*0.4     // 0.2 to 0.6 intensity
		effect.Duration = 600 + power*1800 // 10 to 40 minutes
		effect.MetamorphOrder = 1          // Always low order, but chaotic
		effect.MetamorphArea = 5.0 + power*15.0

	case "spawn_hostile":
		effect.Target = "entity"
		effect.Description = "The ritual summons hostile entities"
		effect.Value = 0.3 + power*0.5     // 0.3 to 0.8 strength
		effect.Duration = 900 + power*1800 // 15 to 45 minutes
		effect.SpawnEntityType = "hostile_spirit"
		effect.SpawnCount = 1 + int(power*2) // 1 to 3 entities
		effect.SpawnRadius = 3.0 + power*7.0
	}

	return effect
}

// checkRitualLocation checks if a location is valid for a ritual
func checkRitualLocation(requiredLocation string, position ecs.Vector3, world *ecs.World) bool {
	// Check environment entities in the vicinity
	entities := world.GetEntitiesWithTag("environment")

	// Search radius
	radius := 10.0

	for _, entity := range entities {
		// Get position
		transformComp, has := entity.GetComponent(ecs.TransformComponentID)
		if !has {
			continue
		}

		transform := transformComp.(*ecs.TransformComponent)
		distance := position.Distance(transform.Position)

		// Check if entity is within radius
		if distance <= radius {
			// Check if entity has required tags
			if entity.HasTag(requiredLocation) {
				return true
			}
		}
	}

	return false
}

// containsString checks if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// uniqueStrings returns a slice with duplicate strings removed
func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	unique := make([]string, 0, len(slice))

	for _, s := range slice {
		if _, exists := keys[s]; !exists {
			keys[s] = true
			unique = append(unique, s)
		}
	}

	return unique
}
