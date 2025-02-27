package fear

import (
	"math"
	"math/rand"
	"time"
)

// BehaviorType represents a type of player behavior
type BehaviorType string

const (
	BehaviorExploring   BehaviorType = "exploring"
	BehaviorHiding      BehaviorType = "hiding"
	BehaviorRunning     BehaviorType = "running"
	BehaviorFighting    BehaviorType = "fighting"
	BehaviorCrafting    BehaviorType = "crafting"
	BehaviorRitualizing BehaviorType = "ritualizing"
)

// FearEvent represents a potential fear-inducing event
type FearEvent struct {
	ID             string
	Type           string
	Intensity      float64        // 0.0-1.0
	Duration       int            // seconds
	CooldownTime   int            // seconds between uses
	LastUsed       int64          // timestamp
	EffectiveAreas []BehaviorType // what player behaviors this is effective against
}

// FearTrigger represents a condition that can trigger fear events
type FearTrigger struct {
	ID          string
	Condition   string      // "time", "location", "action", "state"
	Value       interface{} // depends on condition type
	Events      []string    // potential events to trigger
	Probability float64     // base probability
}

// BehaviorProfile tracks the player's behavior patterns
type BehaviorProfile struct {
	BehaviorCounts     map[BehaviorType]int
	RecentBehaviors    []BehaviorType
	FearResponses      map[string]float64 // event ID -> effectiveness
	PreferredSafeZones []string
	AvoidedAreas       []string
	PlayStyle          map[string]float64 // characteristics like "cautious", "bold"
}

// FearDirector manages the adaptive horror system
type FearDirector struct {
	Events           map[string]*FearEvent
	Triggers         map[string]*FearTrigger
	PlayerProfile    *BehaviorProfile
	CurrentTension   float64 // 0.0-1.0
	SuspenseBuildup  float64 // rate of tension increase
	LastEventTime    int64
	MinEventSpacing  int // minimum seconds between events
	SuccessfulEvents []string
	FailedEvents     []string
	rng              *rand.Rand
}

// NewFearDirector creates a new FearDirector
func NewFearDirector() *FearDirector {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return &FearDirector{
		Events:          make(map[string]*FearEvent),
		Triggers:        make(map[string]*FearTrigger),
		PlayerProfile:   newBehaviorProfile(),
		CurrentTension:  0.2, // start with a little tension
		SuspenseBuildup: 0.01,
		MinEventSpacing: 120, // 2 minutes minimum between events
		rng:             rng,
	}
}

// Initialize sets up the fear system with initial events and triggers
func (fd *FearDirector) Initialize(worldSeed int64) {
	// Reset the RNG with the world seed for consistency
	fd.rng = rand.New(rand.NewSource(worldSeed))

	// Generate base fear events
	fd.generateBaseEvents()

	// Generate base triggers
	fd.generateBaseTriggers()
}

// generateBaseEvents creates the initial set of fear events
func (fd *FearDirector) generateBaseEvents() {
	// Ambient events
	fd.Events["distant_howl"] = &FearEvent{
		ID:             "distant_howl",
		Type:           "ambient_sound",
		Intensity:      0.3,
		Duration:       5,
		CooldownTime:   300,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorHiding},
	}

	fd.Events["rustling_bushes"] = &FearEvent{
		ID:             "rustling_bushes",
		Type:           "ambient_movement",
		Intensity:      0.4,
		Duration:       8,
		CooldownTime:   240,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorCrafting},
	}

	fd.Events["shadow_movement"] = &FearEvent{
		ID:             "shadow_movement",
		Type:           "visual_peripheral",
		Intensity:      0.5,
		Duration:       3,
		CooldownTime:   420,
		EffectiveAreas: []BehaviorType{BehaviorHiding, BehaviorCrafting, BehaviorRitualizing},
	}

	// Environmental events
	fd.Events["sudden_fog"] = &FearEvent{
		ID:             "sudden_fog",
		Type:           "environmental",
		Intensity:      0.6,
		Duration:       300,
		CooldownTime:   900,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorRunning},
	}

	fd.Events["temperature_drop"] = &FearEvent{
		ID:             "temperature_drop",
		Type:           "environmental",
		Intensity:      0.4,
		Duration:       240,
		CooldownTime:   720,
		EffectiveAreas: []BehaviorType{BehaviorHiding, BehaviorCrafting},
	}

	// Direct threats
	fd.Events["creature_stalking"] = &FearEvent{
		ID:             "creature_stalking",
		Type:           "threat",
		Intensity:      0.7,
		Duration:       180,
		CooldownTime:   1200,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorHiding, BehaviorCrafting},
	}

	fd.Events["sudden_attack"] = &FearEvent{
		ID:             "sudden_attack",
		Type:           "threat",
		Intensity:      0.9,
		Duration:       30,
		CooldownTime:   1800,
		EffectiveAreas: []BehaviorType{BehaviorHiding, BehaviorCrafting, BehaviorRitualizing},
	}

	// Strange occurrences
	fd.Events["item_movement"] = &FearEvent{
		ID:             "item_movement",
		Type:           "paranormal",
		Intensity:      0.5,
		Duration:       10,
		CooldownTime:   600,
		EffectiveAreas: []BehaviorType{BehaviorCrafting, BehaviorRitualizing},
	}

	fd.Events["whispers"] = &FearEvent{
		ID:             "whispers",
		Type:           "paranormal",
		Intensity:      0.6,
		Duration:       20,
		CooldownTime:   540,
		EffectiveAreas: []BehaviorType{BehaviorHiding, BehaviorRitualizing},
	}

	// Reality distortions
	fd.Events["visual_glitch"] = &FearEvent{
		ID:             "visual_glitch",
		Type:           "distortion",
		Intensity:      0.7,
		Duration:       8,
		CooldownTime:   780,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorRunning, BehaviorFighting},
	}

	fd.Events["gravity_shift"] = &FearEvent{
		ID:             "gravity_shift",
		Type:           "distortion",
		Intensity:      0.8,
		Duration:       15,
		CooldownTime:   1500,
		EffectiveAreas: []BehaviorType{BehaviorExploring, BehaviorRunning, BehaviorFighting},
	}
}

// generateBaseTriggers creates the initial set of fear triggers
func (fd *FearDirector) generateBaseTriggers() {
	// Time-based triggers
	fd.Triggers["night_time"] = &FearTrigger{
		ID:          "night_time",
		Condition:   "time",
		Value:       "night",
		Events:      []string{"distant_howl", "shadow_movement", "whispers", "sudden_fog"},
		Probability: 0.6,
	}

	fd.Triggers["dusk"] = &FearTrigger{
		ID:          "dusk",
		Condition:   "time",
		Value:       "dusk",
		Events:      []string{"temperature_drop", "rustling_bushes"},
		Probability: 0.4,
	}

	// Location-based triggers
	fd.Triggers["dense_forest"] = &FearTrigger{
		ID:          "dense_forest",
		Condition:   "location",
		Value:       "forest",
		Events:      []string{"rustling_bushes", "shadow_movement", "creature_stalking"},
		Probability: 0.5,
	}

	fd.Triggers["clearing"] = &FearTrigger{
		ID:          "clearing",
		Condition:   "location",
		Value:       "clearing",
		Events:      []string{"distant_howl", "sudden_fog", "visual_glitch"},
		Probability: 0.3,
	}

	fd.Triggers["near_water"] = &FearTrigger{
		ID:          "near_water",
		Condition:   "location",
		Value:       "water",
		Events:      []string{"whispers", "gravity_shift", "temperature_drop"},
		Probability: 0.4,
	}

	// Action-based triggers
	fd.Triggers["using_fire"] = &FearTrigger{
		ID:          "using_fire",
		Condition:   "action",
		Value:       "fire",
		Events:      []string{"shadow_movement", "rustling_bushes", "creature_stalking"},
		Probability: 0.5,
	}

	fd.Triggers["ritual_performance"] = &FearTrigger{
		ID:          "ritual_performance",
		Condition:   "action",
		Value:       "ritual",
		Events:      []string{"whispers", "visual_glitch", "gravity_shift", "item_movement"},
		Probability: 0.7,
	}

	// State-based triggers
	fd.Triggers["low_health"] = &FearTrigger{
		ID:          "low_health",
		Condition:   "state",
		Value:       "health_low",
		Events:      []string{"distant_howl", "whispers", "visual_glitch"},
		Probability: 0.6,
	}

	fd.Triggers["in_darkness"] = &FearTrigger{
		ID:          "in_darkness",
		Condition:   "state",
		Value:       "darkness",
		Events:      []string{"rustling_bushes", "shadow_movement", "whispers"},
		Probability: 0.7,
	}
}

// Update should be called each game tick to update the fear system state
func (fd *FearDirector) Update(deltaTime float64, playerState map[string]interface{}) *FearEvent {
	// Extract player information
	position, _ := playerState["position"].([3]float64)
	behavior, _ := playerState["current_behavior"].(string)
	currentTime := time.Now().Unix()

	// Update tension based on time
	fd.CurrentTension += fd.SuspenseBuildup * deltaTime
	if fd.CurrentTension > 1.0 {
		fd.CurrentTension = 1.0
	}

	// Update player behavior profile
	fd.updateBehaviorProfile(BehaviorType(behavior), position)

	// Check if enough time has passed since the last event
	if currentTime-fd.LastEventTime < int64(fd.MinEventSpacing) {
		return nil
	}

	// Calculate base probability of an event happening based on tension
	eventProbability := fd.CurrentTension * 0.1 * deltaTime

	// Check if we should trigger an event
	if fd.rng.Float64() < eventProbability {
		// Find active triggers
		activeTriggers := fd.findActiveTriggers(playerState)

		if len(activeTriggers) > 0 {
			// Pick a trigger
			selectedTrigger := activeTriggers[fd.rng.Intn(len(activeTriggers))]

			// Find viable events from this trigger
			viableEvents := fd.findViableEvents(selectedTrigger, BehaviorType(behavior), currentTime)

			if len(viableEvents) > 0 {
				// Select event
				selectedEvent := fd.selectOptimalEvent(viableEvents, BehaviorType(behavior))

				// Update event state
				selectedEvent.LastUsed = currentTime
				fd.LastEventTime = currentTime

				// Reset tension partially
				fd.CurrentTension *= 0.7

				return selectedEvent
			}
		}
	}

	return nil
}

// findActiveTriggers returns all triggers that are currently active
func (fd *FearDirector) findActiveTriggers(playerState map[string]interface{}) []*FearTrigger {
	var activeTriggers []*FearTrigger

	for _, trigger := range fd.Triggers {
		isActive := false

		switch trigger.Condition {
		case "time":
			timeOfDay, _ := playerState["time_of_day"].(string)
			isActive = timeOfDay == trigger.Value

		case "location":
			location, _ := playerState["biome"].(string)
			isActive = location == trigger.Value

		case "action":
			action, _ := playerState["current_action"].(string)
			isActive = action == trigger.Value

		case "state":
			switch trigger.Value {
			case "health_low":
				health, _ := playerState["health"].(float64)
				isActive = health < 0.3

			case "darkness":
				lightLevel, _ := playerState["light_level"].(float64)
				isActive = lightLevel < 0.2

			default:
				state, exists := playerState[trigger.Value.(string)]
				if exists {
					boolState, isBool := state.(bool)
					if isBool {
						isActive = boolState
					}
				}
			}
		}

		// Apply probability
		if isActive && fd.rng.Float64() < trigger.Probability {
			activeTriggers = append(activeTriggers, trigger)
		}
	}

	return activeTriggers
}

// findViableEvents returns events from the trigger that are currently viable
func (fd *FearDirector) findViableEvents(trigger *FearTrigger, currentBehavior BehaviorType, currentTime int64) []*FearEvent {
	var viableEvents []*FearEvent

	for _, eventID := range trigger.Events {
		event, exists := fd.Events[eventID]
		if !exists {
			continue
		}

		// Check cooldown
		if currentTime-event.LastUsed < int64(event.CooldownTime) {
			continue
		}

		// Check if effective against current behavior
		isEffective := false
		for _, behavior := range event.EffectiveAreas {
			if behavior == currentBehavior {
				isEffective = true
				break
			}
		}

		if isEffective {
			viableEvents = append(viableEvents, event)
		}
	}

	return viableEvents
}

// selectOptimalEvent chooses the most effective event
func (fd *FearDirector) selectOptimalEvent(events []*FearEvent, currentBehavior BehaviorType) *FearEvent {
	if len(events) == 1 {
		return events[0]
	}

	var bestEvent *FearEvent
	var highestScore float64 = -1

	for _, event := range events {
		// Base score is the event intensity
		score := event.Intensity

		// Adjust based on past effectiveness
		effectiveness, found := fd.PlayerProfile.FearResponses[event.ID]
		if found {
			score *= (0.5 + effectiveness)
		}

		// Give preference to events that haven't been used recently
		var timeFactor float64 = 1.0
		if event.LastUsed > 0 {
			timeElapsed := float64(time.Now().Unix() - event.LastUsed)
			timeFactor = math.Min(2.0, 1.0+(timeElapsed/float64(event.CooldownTime*2)))
		}
		score *= timeFactor

		// Add some randomness
		score *= (0.9 + fd.rng.Float64()*0.2)

		if score > highestScore {
			highestScore = score
			bestEvent = event
		}
	}

	return bestEvent
}

// RecordEventEffectiveness records how effective a fear event was
func (fd *FearDirector) RecordEventEffectiveness(eventID string, effectiveness float64) {
	if _, exists := fd.Events[eventID]; !exists {
		return
	}

	// Record the effectiveness
	currentValue, found := fd.PlayerProfile.FearResponses[eventID]
	if found {
		// Weighted average, giving more weight to the new value
		fd.PlayerProfile.FearResponses[eventID] = currentValue*0.7 + effectiveness*0.3
	} else {
		fd.PlayerProfile.FearResponses[eventID] = effectiveness
	}

	// Track successful and failed events
	if effectiveness > 0.6 {
		fd.SuccessfulEvents = append(fd.SuccessfulEvents, eventID)

		// Adjust suspense buildup for this event type
		fd.SuspenseBuildup *= 1.02 // Gradually increase tension buildup rate
	} else if effectiveness < 0.3 {
		fd.FailedEvents = append(fd.FailedEvents, eventID)

		// Adjust for failed event
		fd.SuspenseBuildup *= 0.98 // Decrease tension buildup slightly
	}

	// Ensure suspense buildup stays in reasonable bounds
	if fd.SuspenseBuildup < 0.002 {
		fd.SuspenseBuildup = 0.002
	} else if fd.SuspenseBuildup > 0.05 {
		fd.SuspenseBuildup = 0.05
	}
}

// RecordPlayerReaction records how the player reacted to a situation
func (fd *FearDirector) RecordPlayerReaction(eventID string, reaction BehaviorType) {
	// Record the reaction for learning
	fd.PlayerProfile.RecentBehaviors = append(fd.PlayerProfile.RecentBehaviors, reaction)

	// Keep only the last 20 behaviors
	if len(fd.PlayerProfile.RecentBehaviors) > 20 {
		fd.PlayerProfile.RecentBehaviors = fd.PlayerProfile.RecentBehaviors[1:]
	}

	// Update behavior counts
	fd.PlayerProfile.BehaviorCounts[reaction]++

	// Update event effective areas based on reaction
	event, exists := fd.Events[eventID]
	if exists {
		// If the player responds with fear (running/hiding), mark this as effective
		if reaction == BehaviorRunning || reaction == BehaviorHiding {
			fd.RecordEventEffectiveness(eventID, 0.8)

			// Ensure this behavior is marked as effective for this event
			hasEffectiveArea := false
			for _, behavior := range event.EffectiveAreas {
				if behavior == reaction {
					hasEffectiveArea = true
					break
				}
			}

			if !hasEffectiveArea {
				event.EffectiveAreas = append(event.EffectiveAreas, reaction)
			}
		} else if reaction == BehaviorFighting {
			// If the player fights, this was less effective
			fd.RecordEventEffectiveness(eventID, 0.3)
		}
	}
}

// IdentifySafeZone marks an area as a safe zone for the player
func (fd *FearDirector) IdentifySafeZone(areaID string) {
	// Check if already known
	for _, zone := range fd.PlayerProfile.PreferredSafeZones {
		if zone == areaID {
			return
		}
	}

	fd.PlayerProfile.PreferredSafeZones = append(fd.PlayerProfile.PreferredSafeZones, areaID)
}

// IdentifyAvoidedArea marks an area as avoided by the player
func (fd *FearDirector) IdentifyAvoidedArea(areaID string) {
	// Check if already known
	for _, area := range fd.PlayerProfile.AvoidedAreas {
		if area == areaID {
			return
		}
	}

	fd.PlayerProfile.AvoidedAreas = append(fd.PlayerProfile.AvoidedAreas, areaID)
}

// UpdatePlayStyle adjusts the player's playstyle profile
func (fd *FearDirector) UpdatePlayStyle(behavior BehaviorType) {
	// Update play style based on behavior patterns

	// Calculate total behavior count
	total := 0
	for _, count := range fd.PlayerProfile.BehaviorCounts {
		total += count
	}

	if total < 10 {
		return // Not enough data yet
	}

	// Calculate proportions of behaviors
	exploringPct := float64(fd.PlayerProfile.BehaviorCounts[BehaviorExploring]) / float64(total)
	hidingPct := float64(fd.PlayerProfile.BehaviorCounts[BehaviorHiding]) / float64(total)
	runningPct := float64(fd.PlayerProfile.BehaviorCounts[BehaviorRunning]) / float64(total)
	fightingPct := float64(fd.PlayerProfile.BehaviorCounts[BehaviorFighting]) / float64(total)

	// Update play style values
	fd.PlayerProfile.PlayStyle["cautious"] = hidingPct*0.7 + runningPct*0.3
	fd.PlayerProfile.PlayStyle["bold"] = exploringPct*0.5 + fightingPct*0.5
	fd.PlayerProfile.PlayStyle["methodical"] = (1.0 - runningPct) * 0.8
	fd.PlayerProfile.PlayStyle["impulsive"] = runningPct*0.6 + fightingPct*0.4
}

// SaveState returns a serializable representation of the fear director state
func (fd *FearDirector) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"events":           fd.Events,
		"triggers":         fd.Triggers,
		"playerProfile":    fd.PlayerProfile,
		"currentTension":   fd.CurrentTension,
		"suspenseBuildup":  fd.SuspenseBuildup,
		"lastEventTime":    fd.LastEventTime,
		"minEventSpacing":  fd.MinEventSpacing,
		"successfulEvents": fd.SuccessfulEvents,
		"failedEvents":     fd.FailedEvents,
	}
}

// LoadState initializes the fear director from a saved state
func (fd *FearDirector) LoadState(state map[string]interface{}) {
	if events, ok := state["events"].(map[string]*FearEvent); ok {
		fd.Events = events
	}

	if triggers, ok := state["triggers"].(map[string]*FearTrigger); ok {
		fd.Triggers = triggers
	}

	if profile, ok := state["playerProfile"].(*BehaviorProfile); ok {
		fd.PlayerProfile = profile
	}

	if tension, ok := state["currentTension"].(float64); ok {
		fd.CurrentTension = tension
	}

	if buildup, ok := state["suspenseBuildup"].(float64); ok {
		fd.SuspenseBuildup = buildup
	}

	if lastTime, ok := state["lastEventTime"].(int64); ok {
		fd.LastEventTime = lastTime
	}

	if spacing, ok := state["minEventSpacing"].(int); ok {
		fd.MinEventSpacing = spacing
	}

	if successful, ok := state["successfulEvents"].([]string); ok {
		fd.SuccessfulEvents = successful
	}

	if failed, ok := state["failedEvents"].([]string); ok {
		fd.FailedEvents = failed
	}
}

// Helper functions

// updateBehaviorProfile updates the player's behavior profile
func (fd *FearDirector) updateBehaviorProfile(behavior BehaviorType, position [3]float64) {
	// Check if this is a new behavior type
	if _, exists := fd.PlayerProfile.BehaviorCounts[behavior]; !exists {
		fd.PlayerProfile.BehaviorCounts[behavior] = 0
	}

	// Add current behavior to the list of recent behaviors
	fd.PlayerProfile.RecentBehaviors = append(fd.PlayerProfile.RecentBehaviors, behavior)
	if len(fd.PlayerProfile.RecentBehaviors) > 20 {
		fd.PlayerProfile.RecentBehaviors = fd.PlayerProfile.RecentBehaviors[1:]
	}

	// Increment the behavior count
	fd.PlayerProfile.BehaviorCounts[behavior]++

	// Update play style
	fd.UpdatePlayStyle(behavior)
}

// newBehaviorProfile creates a new empty behavior profile
func newBehaviorProfile() *BehaviorProfile {
	return &BehaviorProfile{
		BehaviorCounts:     make(map[BehaviorType]int),
		RecentBehaviors:    make([]BehaviorType, 0),
		FearResponses:      make(map[string]float64),
		PreferredSafeZones: make([]string, 0),
		AvoidedAreas:       make([]string, 0),
		PlayStyle:          make(map[string]float64),
	}
}
