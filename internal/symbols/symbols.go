package symbols

import (
	"math/rand"
	"time"
)

// Symbol represents a discoverable occult symbol in the game world
type Symbol struct {
	ID                string
	Category          string
	VisualComplexity  float64
	DiscoveryTime     int64
	DiscoveryLocation [3]float64
	PlayerKnowledge   float64
	RelatedSymbols    []string
	MeaningComponents []string
	VisualData        string
}

// Ritual represents a ritual that can be performed with specific symbols and components
type Ritual struct {
	ID              string
	Name            string
	DiscoveryState  string // "unknown", "partial", "complete"
	RequiredSymbols []string
	Components      []RitualComponent
	Effects         []RitualEffect
	MasteryLevel    float64
	LastPerformed   int64
	SuccessCount    int
	FailureCount    int
	EvolutionPath   []string
}

// RitualComponent represents a component needed for a ritual
type RitualComponent struct {
	Type       string // "item", "location", "action"
	ID         string
	Properties []string
}

// RitualEffect represents an effect that a ritual can produce
type RitualEffect struct {
	Type     string
	ID       string
	Duration int
	Order    int // For metamorphosis effects
	Category string
}

// SymbolManager handles all symbol and ritual-related functionality
type SymbolManager struct {
	Symbols           map[string]*Symbol
	Rituals           map[string]*Ritual
	DiscoveredSymbols map[string]*Symbol
	KnownRituals      map[string]*Ritual
	rng               *rand.Rand
}

// NewSymbolManager creates a new SymbolManager instance
func NewSymbolManager() *SymbolManager {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return &SymbolManager{
		Symbols:           make(map[string]*Symbol),
		Rituals:           make(map[string]*Ritual),
		DiscoveredSymbols: make(map[string]*Symbol),
		KnownRituals:      make(map[string]*Ritual),
		rng:               rng,
	}
}

// Initialize sets up the initial symbols and rituals pool
func (sm *SymbolManager) Initialize(worldSeed int64) {
	// Reset the RNG with the world seed for consistency
	sm.rng = rand.New(rand.NewSource(worldSeed))

	// Generate base symbols
	sm.generateBaseSymbols()

	// Generate initial rituals
	sm.generateBaseRituals()
}

// generateBaseSymbols creates the initial set of symbols
func (sm *SymbolManager) generateBaseSymbols() {
	// Define categories
	categories := []string{"elder", "nature", "void", "transformation", "perception"}

	// Generate a set of base symbols for each category
	for _, category := range categories {
		for i := 0; i < 3+sm.rng.Intn(5); i++ {
			symbolID := "s_" + category + "_" + generateID(sm.rng, 5)

			// Create a new symbol
			symbol := &Symbol{
				ID:                symbolID,
				Category:          category,
				VisualComplexity:  0.3 + sm.rng.Float64()*0.7,
				PlayerKnowledge:   0.0,
				MeaningComponents: []string{category},
				VisualData:        "symbols/" + symbolID + ".png",
			}

			// Add it to the symbols pool
			sm.Symbols[symbolID] = symbol
		}
	}

	// Connect related symbols
	for id, symbol := range sm.Symbols {
		// Each symbol can be related to 0-3 other symbols
		relatedCount := sm.rng.Intn(4)
		for i := 0; i < relatedCount; i++ {
			// Pick a random symbol that isn't this one
			var relatedID string
			for {
				randomIdx := sm.rng.Intn(len(sm.Symbols))
				j := 0
				for key := range sm.Symbols {
					if j == randomIdx {
						relatedID = key
						break
					}
					j++
				}

				if relatedID != id && !contains(symbol.RelatedSymbols, relatedID) {
					break
				}
			}

			// Add the relation
			symbol.RelatedSymbols = append(symbol.RelatedSymbols, relatedID)
		}
	}
}

// generateBaseRituals creates the initial set of rituals
func (sm *SymbolManager) generateBaseRituals() {
	// Define basic ritual types
	ritualTypes := []string{
		"perception_shift",
		"reality_warp",
		"summoning",
		"banishing",
		"protection",
	}

	// Generate rituals for each type
	for _, ritualType := range ritualTypes {
		for i := 0; i < 1+sm.rng.Intn(3); i++ {
			ritualID := "r_" + ritualType + "_" + generateID(sm.rng, 5)

			// Pick 1-3 required symbols
			var requiredSymbols []string
			symbolCount := 1 + sm.rng.Intn(3)
			categories := []string{"elder", "nature", "void", "transformation", "perception"}
			preferredCategory := categories[sm.rng.Intn(len(categories))]

			// Try to get symbols from the preferred category
			var possibleSymbols []string
			for id, symbol := range sm.Symbols {
				if symbol.Category == preferredCategory {
					possibleSymbols = append(possibleSymbols, id)
				}
			}

			// If not enough from preferred, add from any category
			if len(possibleSymbols) < symbolCount {
				for id := range sm.Symbols {
					if !contains(possibleSymbols, id) {
						possibleSymbols = append(possibleSymbols, id)
					}
				}
			}

			// Randomly select required symbols
			for i := 0; i < symbolCount && i < len(possibleSymbols); i++ {
				idx := sm.rng.Intn(len(possibleSymbols))
				requiredSymbols = append(requiredSymbols, possibleSymbols[idx])
				// Remove to avoid duplicates
				possibleSymbols = append(possibleSymbols[:idx], possibleSymbols[idx+1:]...)
			}

			// Create components
			components := generateRitualComponents(sm.rng, ritualType)

			// Create effects
			effects := generateRitualEffects(sm.rng, ritualType)

			// Create the ritual
			ritual := &Ritual{
				ID:              ritualID,
				Name:            generateRitualName(sm.rng, ritualType),
				DiscoveryState:  "unknown",
				RequiredSymbols: requiredSymbols,
				Components:      components,
				Effects:         effects,
				MasteryLevel:    0.0,
				SuccessCount:    0,
				FailureCount:    0,
			}

			// Add it to the rituals pool
			sm.Rituals[ritualID] = ritual
		}
	}
}

// DiscoverSymbol marks a symbol as discovered by the player
func (sm *SymbolManager) DiscoverSymbol(symbolID string, location [3]float64) bool {
	symbol, exists := sm.Symbols[symbolID]
	if !exists {
		return false
	}

	// Check if already discovered
	if _, found := sm.DiscoveredSymbols[symbolID]; found {
		return false
	}

	// Mark as discovered
	symbol.DiscoveryTime = time.Now().Unix()
	symbol.DiscoveryLocation = location
	symbol.PlayerKnowledge = 0.1 // Initial knowledge
	sm.DiscoveredSymbols[symbolID] = symbol

	return true
}

// IncreaseSymbolKnowledge increases the player's knowledge about a symbol
func (sm *SymbolManager) IncreaseSymbolKnowledge(symbolID string, amount float64) bool {
	symbol, exists := sm.DiscoveredSymbols[symbolID]
	if !exists {
		return false
	}

	// Increase knowledge, capped at 1.0
	symbol.PlayerKnowledge += amount
	if symbol.PlayerKnowledge > 1.0 {
		symbol.PlayerKnowledge = 1.0
	}

	// Check if this reveals any rituals
	sm.checkRitualDiscovery()

	return true
}

// checkRitualDiscovery checks if any new rituals should be discovered
func (sm *SymbolManager) checkRitualDiscovery() {
	for ritualID, ritual := range sm.Rituals {
		// Skip already discovered
		if _, found := sm.KnownRituals[ritualID]; found {
			continue
		}

		// Check if all required symbols are known with sufficient knowledge
		knownCount := 0
		totalKnowledge := 0.0

		for _, symbolID := range ritual.RequiredSymbols {
			if symbol, found := sm.DiscoveredSymbols[symbolID]; found {
				knownCount++
				totalKnowledge += symbol.PlayerKnowledge
			}
		}

		// Calculate average knowledge
		avgKnowledge := 0.0
		if len(ritual.RequiredSymbols) > 0 {
			avgKnowledge = totalKnowledge / float64(len(ritual.RequiredSymbols))
		}

		// Update discovery state
		if knownCount == len(ritual.RequiredSymbols) {
			if avgKnowledge > 0.8 {
				// Fully discovered
				ritual.DiscoveryState = "complete"
				sm.KnownRituals[ritualID] = ritual
			} else if avgKnowledge > 0.4 {
				// Partially discovered
				ritual.DiscoveryState = "partial"
				sm.KnownRituals[ritualID] = ritual
			}
		}
	}
}

// PerformRitual attempts to perform a ritual and returns the effects
func (sm *SymbolManager) PerformRitual(ritualID string, location [3]float64, components []string) ([]RitualEffect, bool) {
	ritual, exists := sm.KnownRituals[ritualID]
	if !exists || ritual.DiscoveryState == "unknown" {
		return nil, false
	}

	// Check if components are correct
	componentSuccess := true
	if ritual.DiscoveryState == "complete" {
		// Need exact components
		// This is simplified - would need more complex matching in a real implementation
		if len(components) != len(ritual.Components) {
			componentSuccess = false
		}
	}

	// Record attempt
	ritual.LastPerformed = time.Now().Unix()

	if componentSuccess {
		// Success
		ritual.SuccessCount++
		ritual.MasteryLevel += 0.1
		if ritual.MasteryLevel > 1.0 {
			ritual.MasteryLevel = 1.0
		}
		return ritual.Effects, true
	} else {
		// Failure
		ritual.FailureCount++
		return nil, false
	}
}

// GetSymbolsInRange returns symbols located within a given range of a point
func (sm *SymbolManager) GetSymbolsInRange(position [3]float64, radius float64) []*Symbol {
	var result []*Symbol

	for _, symbol := range sm.Symbols {
		// Only include if it has a discovery location (placed in the world)
		if symbol.DiscoveryTime == 0 {
			continue
		}

		// Calculate distance
		dx := position[0] - symbol.DiscoveryLocation[0]
		dy := position[1] - symbol.DiscoveryLocation[1]
		dz := position[2] - symbol.DiscoveryLocation[2]
		distSquared := dx*dx + dy*dy + dz*dz

		if distSquared <= radius*radius {
			result = append(result, symbol)
		}
	}

	return result
}

// SaveState returns a serializable representation of the symbol manager state
func (sm *SymbolManager) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"symbols":           sm.Symbols,
		"rituals":           sm.Rituals,
		"discoveredSymbols": sm.DiscoveredSymbols,
		"knownRituals":      sm.KnownRituals,
	}
}

// LoadState initializes the symbol manager from a saved state
func (sm *SymbolManager) LoadState(state map[string]interface{}) {
	if symbols, ok := state["symbols"].(map[string]*Symbol); ok {
		sm.Symbols = symbols
	}

	if rituals, ok := state["rituals"].(map[string]*Ritual); ok {
		sm.Rituals = rituals
	}

	if discovered, ok := state["discoveredSymbols"].(map[string]*Symbol); ok {
		sm.DiscoveredSymbols = discovered
	}

	if known, ok := state["knownRituals"].(map[string]*Ritual); ok {
		sm.KnownRituals = known
	}
}

// Helper functions
func generateID(rng *rand.Rand, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateRitualComponents(rng *rand.Rand, ritualType string) []RitualComponent {
	components := []RitualComponent{}

	// Add 1-3 item components
	itemCount := 1 + rng.Intn(3)
	possibleItems := []string{"candle", "bone", "herb", "crystal", "blood", "feather", "moth_eyes", "ash", "water"}
	for i := 0; i < itemCount; i++ {
		if len(possibleItems) > 0 {
			idx := rng.Intn(len(possibleItems))
			components = append(components, RitualComponent{
				Type: "item",
				ID:   possibleItems[idx],
			})
			// Remove to avoid duplicates
			possibleItems = append(possibleItems[:idx], possibleItems[idx+1:]...)
		}
	}

	// Add location requirements
	locationProbability := rng.Float64()
	if locationProbability > 0.3 {
		locationProps := []string{"water", "dark", "elevated", "clearing", "rock", "tree", "moonlight"}
		propertyCount := 1 + rng.Intn(2)
		var properties []string

		for i := 0; i < propertyCount; i++ {
			if len(locationProps) > 0 {
				idx := rng.Intn(len(locationProps))
				properties = append(properties, locationProps[idx])
				// Remove to avoid duplicates
				locationProps = append(locationProps[:idx], locationProps[idx+1:]...)
			}
		}

		components = append(components, RitualComponent{
			Type:       "location",
			Properties: properties,
		})
	}

	// Add action requirements
	actionProbability := rng.Float64()
	if actionProbability > 0.2 {
		possibleActions := []string{"circle", "chant", "burn", "bury", "mark", "sacrifice", "meditate"}
		actionCount := 1 + rng.Intn(3)
		var sequence []string

		for i := 0; i < actionCount; i++ {
			if len(possibleActions) > 0 {
				idx := rng.Intn(len(possibleActions))
				sequence = append(sequence, possibleActions[idx])
				// Remove to avoid duplicates
				possibleActions = append(possibleActions[:idx], possibleActions[idx+1:]...)
			}
		}

		components = append(components, RitualComponent{
			Type:       "action",
			Properties: sequence,
		})
	}

	return components
}

func generateRitualEffects(rng *rand.Rand, ritualType string) []RitualEffect {
	effects := []RitualEffect{}

	// Primary effect based on ritual type
	primaryEffect := RitualEffect{
		Type:     "metamorphosis",
		Duration: 300 + rng.Intn(3600), // 5 min to 1 hour
	}

	// Set order and category based on type
	switch ritualType {
	case "perception_shift":
		primaryEffect.Order = 1 + rng.Intn(3) // 1-3
		primaryEffect.Category = "perception_shift"
	case "reality_warp":
		primaryEffect.Order = 2 + rng.Intn(3) // 2-4
		primaryEffect.Category = "reality_warp"
	case "summoning":
		primaryEffect.Order = 3 + rng.Intn(2) // 3-4
		primaryEffect.Category = "entity_manifestation"
	case "banishing":
		primaryEffect.Order = 2 + rng.Intn(3) // 2-4
		primaryEffect.Category = "purification"
	case "protection":
		primaryEffect.Order = 1 + rng.Intn(2) // 1-2
		primaryEffect.Category = "barrier"
	default:
		primaryEffect.Order = 1 + rng.Intn(5) // 1-5
		primaryEffect.Category = "unknown"
	}

	effects = append(effects, primaryEffect)

	// Chance for secondary effect
	if rng.Float64() > 0.7 {
		secondaryEffects := []string{"night_vision", "temperature_control", "stealth", "detection", "strength"}
		effects = append(effects, RitualEffect{
			Type:     "player_ability",
			ID:       secondaryEffects[rng.Intn(len(secondaryEffects))],
			Duration: 600 + rng.Intn(1800), // 10-30 min
		})
	}

	return effects
}

func generateRitualName(rng *rand.Rand, ritualType string) string {
	prefixes := map[string][]string{
		"perception_shift": {"Veil", "Eye", "Mind", "Sight", "Vision"},
		"reality_warp":     {"Bend", "Warp", "Twist", "Fold", "Break"},
		"summoning":        {"Call", "Summon", "Beckon", "Invoke", "Manifest"},
		"banishing":        {"Banish", "Cleanse", "Purge", "Expel", "Ward"},
		"protection":       {"Shield", "Guard", "Protect", "Shelter", "Barrier"},
	}

	suffixes := map[string][]string{
		"perception_shift": {"of Insight", "of Truth", "of Awakening", "of Clarity", "of Revelation"},
		"reality_warp":     {"of Space", "of Matter", "of Dimensions", "of Reality", "of Existence"},
		"summoning":        {"of Shadows", "of Spirits", "of Entities", "of Beings", "of Forces"},
		"banishing":        {"of Purity", "of Banishment", "of Cleansing", "of Freedom", "of Release"},
		"protection":       {"of Safety", "of Defense", "of Sanctuary", "of Warding", "of Shielding"},
	}

	// Get relevant lists or use generic ones
	prefixList, ok := prefixes[ritualType]
	if !ok {
		prefixList = []string{"Ancient", "Dark", "Hidden", "Forgotten", "Secret"}
	}

	suffixList, ok := suffixes[ritualType]
	if !ok {
		suffixList = []string{"Ritual", "Ceremony", "Rite", "Practice", "Invocation"}
	}

	// Generate name
	prefix := prefixList[rng.Intn(len(prefixList))]
	suffix := suffixList[rng.Intn(len(suffixList))]

	return prefix + " " + suffix
}
