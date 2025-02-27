package fear

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"echo-taiga/internal/engine/ecs"
)

// ActionType represents the type of action the player is performing
type ActionType string

const (
	ActionMoving     ActionType = "moving"
	ActionRunning    ActionType = "running"
	ActionHiding     ActionType = "hiding"
	ActionFighting   ActionType = "fighting"
	ActionCrafting   ActionType = "crafting"
	ActionExploring  ActionType = "exploring"
	ActionInspecting ActionType = "inspecting"
	ActionResting    ActionType = "resting"
	ActionNone       ActionType = "none"
)

// EmotionalResponse represents the player's emotional state
type EmotionalResponse int

const (
	EmotionCalm       EmotionalResponse = 0
	EmotionAlert      EmotionalResponse = 1
	EmotionNervous    EmotionalResponse = 2
	EmotionFrightened EmotionalResponse = 3
	EmotionTerrified  EmotionalResponse = 4
)

// PlayerAction represents an action taken by the player
type PlayerAction struct {
	Type           ActionType        // Type of action
	Timestamp      time.Time         // When the action occurred
	Position       ecs.Vector3       // Where the action occurred
	Direction      ecs.Vector3       // Direction player was facing
	Speed          float64           // Movement speed (if moving)
	Target         ecs.EntityID      // Target entity (if any)
	LookingAround  bool              // Whether the player was looking around
	EmotionalState EmotionalResponse // Estimated emotional state
	ContextTags    []string          // Tags describing the context
	AreaType       string            // Type of area (forest, cave, etc.)
	LightLevel     float64           // Light level (0-1)
	TimeOfDay      float64           // Time of day (0-1, 0 = midnight, 0.5 = noon)
}

// BehaviorProfile represents the analyzed behavior patterns of the player
type BehaviorProfile struct {
	// General behavior tendencies (0-1 scales)
	RiskAversion float64 // How risk-averse is the player
	Thoroughness float64 // How thoroughly does player explore
	Caution      float64 // How cautious is the player
	Aggression   float64 // How aggressive is the player
	Patience     float64 // How patient is the player
	Curiosity    float64 // How curious is the player

	// Specific fear triggers and ratings (0-1 scales)
	FearTriggers map[string]float64 // Specific things that scare the player

	// Comfort factors
	ComfortZones map[string]float64 // Areas or situations where player feels safe

	// Movement patterns
	MovementPatterns map[string]float64 // How the player typically moves

	// Decision patterns
	DecisionPatterns map[string]float64 // How the player makes decisions

	// Time-based patterns
	TimePatterns map[string]float64 // Time-based behavior patterns

	// Response to scares
	ScareResponses map[string]float64 // How the player responds to scares

	// Adaptability (how quickly player adapts to new situations)
	Adaptability float64

	// Predictability (how predictable player's actions are)
	Predictability float64
}

// ScareEvent represents a scary event created by the system
type ScareEvent struct {
	ID                string       // Unique identifier
	Type              string       // Type of scare
	Subtype           string       // Specific subtype
	Intensity         float64      // How intense is the scare (0-1)
	StartPosition     ecs.Vector3  // Where the scare begins
	TargetPosition    ecs.Vector3  // Where the scare is directed
	Duration          float64      // How long the scare lasts (seconds)
	LightingEffect    string       // Effect on lighting
	SoundEffect       string       // Sound effect to play
	EntityEffect      string       // Effect on entities
	EnvironmentEffect string       // Effect on environment
	EffectRadius      float64      // Radius of effect
	RequiredSetup     []string     // Required setup steps
	Cooldown          float64      // Cooldown before similar scare can be used
	SuccessRating     time.Time    // How successful was this scare (filled after)
	Tags              []string     // Tags describing the scare
	ExclusionTags     []string     // Tags for scares that shouldn't happen close to this
	EntityID          ecs.EntityID // Associated entity (if any)
	MetamorphID       string       // Associated metamorphosis (if any)
}

// TensionLevelName maps tension level to a name
var TensionLevelName = map[int]string{
	0: "Calm",
	1: "Uneasy",
	2: "Tense",
	3: "Frightening",
	4: "Terrifying",
}

// FearProfile represents what scares a specific player
type FearProfile struct {
	// General fear ratings (0-1 scales)
	DarknessFear       float64 // Fear of darkness
	EnclosedSpacesFear float64 // Fear of enclosed spaces
	HeightsFear        float64 // Fear of heights
	WaterFear          float64 // Fear of water
	FireFear           float64 // Fear of fire
	MonstersFear       float64 // Fear of monsters
	InsectsFear        float64 // Fear of insects
	GoreFear           float64 // Fear of blood/gore
	JumpscaresFear     float64 // Fear of jumpscares
	PsychologicalFear  float64 // Fear of psychological horror
	NoiseFear          float64 // Fear of loud noises
	SilenceFear        float64 // Fear of silence
	UnknownFear        float64 // Fear of the unknown
	IsolationFear      float64 // Fear of isolation
	ParanormalFear     float64 // Fear of paranormal

	// Specific entity fears
	EntityFears map[string]float64

	// Environment fears
	EnvironmentFears map[string]float64

	// Contextual fears (situations)
	ContextualFears map[string]float64

	// Most effective scare types
	EffectiveScares map[string]float64

	// Habituation (how quickly player gets used to scares)
	Habituation float64

	// Recovery (how quickly player recovers from scares)
	Recovery float64
}

// ScareOpportunity represents an identified opportunity to scare the player
type ScareOpportunity struct {
	Timestamp      time.Time
	ScareTypes     []string    // Potential scare types
	Position       ecs.Vector3 // Where to trigger the scare
	OptimalTiming  float64     // Best time to trigger (seconds from now)
	EstimatedValue float64     // Estimated effectiveness (0-1)
	PlayerState    string      // What player is doing
	Context        string      // Context for the scare
	TensionLevel   int         // Current tension level (0-4)
}

// Director manages and orchestrates scary events based on player behavior
type Director struct {
	world *ecs.World

	// Player tracking
	playerID       ecs.EntityID
	playerPosition ecs.Vector3
	playerLastSeen time.Time

	// Action history
	actionHistory  []PlayerAction
	maxHistorySize int

	// Behavior profiling
	behaviorProfile *BehaviorProfile
	fearProfile     *FearProfile

	// Scare management
	scareHistory       []ScareEvent
	scareOpportunities []ScareOpportunity
	currentScares      map[string]*ScareEvent
	scareCooldowns     map[string]time.Time

	// Tension curve management
	tensionCurve      float64   // Current tension (0-1)
	targetTension     float64   // Target tension level (0-1)
	tensionDirection  int       // -1 decreasing, 0 stable, 1 increasing
	tensionLevel      int       // Discretized tension level (0-4)
	lastTensionChange time.Time // When tension last changed levels
	tensionPhase      string    // build, peak, release, calm

	// Timing
	lastScareTime    time.Time
	lastAnalysisTime time.Time

	// Templates
	scareTemplates map[string]ScareEvent

	// Environment awareness
	currentAreaType   string
	currentLightLevel float64
	currentTimeOfDay  float64

	// Configuration
	baseScareInterval float64 // Base time between scares (seconds)
	minScareInterval  float64 // Minimum time between scares (seconds)
	maxTensionTime    float64 // Maximum time at high tension (seconds)
	tensionChangeRate float64 // How quickly tension changes (units per second)

	// Adaptive learning
	successfulScares map[string]int
	failedScares     map[string]int

	// Callbacks
	OnScareTriggered func(ScareEvent)

	// Path for saving/loading data
	savePath string

	// Thread safety
	mutex sync.RWMutex
}

// NewDirector creates a new fear director
func NewDirector(world *ecs.World, savePath string) *Director {
	return &Director{
		world:              world,
		actionHistory:      make([]PlayerAction, 0, 100),
		maxHistorySize:     100,
		behaviorProfile:    NewDefaultBehaviorProfile(),
		fearProfile:        NewDefaultFearProfile(),
		scareHistory:       make([]ScareEvent, 0),
		scareOpportunities: make([]ScareOpportunity, 0),
		currentScares:      make(map[string]*ScareEvent),
		scareCooldowns:     make(map[string]time.Time),
		tensionCurve:       0.1,     // Start with low tension
		targetTension:      0.3,     // Initial target is slightly elevated
		tensionDirection:   1,       // Starting by increasing tension
		tensionLevel:       0,       // Start at "calm"
		tensionPhase:       "build", // Start in build phase
		lastTensionChange:  time.Now(),
		baseScareInterval:  180.0, // 3 minutes between major scares by default
		minScareInterval:   60.0,  // Minimum 1 minute between scares
		maxTensionTime:     300.0, // Maximum 5 minutes at high tension
		tensionChangeRate:  0.05,  // Tension units per second
		successfulScares:   make(map[string]int),
		failedScares:       make(map[string]int),
		savePath:           savePath,
		scareTemplates:     make(map[string]ScareEvent),
	}
}

// Initialize sets up the fear director
func (fd *Director) Initialize() error {
	// Try to load existing behavior and fear profiles
	err := fd.LoadProfiles()
	if err != nil {
		// If profiles don't exist, we already have defaults
		fmt.Printf("No existing profiles found, using defaults\n")
	}

	// Load scare templates
	err = fd.LoadScareTemplates()
	if err != nil {
		// Create default templates if they don't exist
		err = fd.CreateDefaultScareTemplates()
		if err != nil {
			return fmt.Errorf("failed to create default scare templates: %v", err)
		}

		// Try loading again
		err = fd.LoadScareTemplates()
		if err != nil {
			return fmt.Errorf("failed to load scare templates: %v", err)
		}
	}

	// Register as a system in the world
	fd.world.AddSystem(fd)

	// Set timestamps
	now := time.Now()
	fd.lastScareTime = now
	fd.lastAnalysisTime = now
	fd.lastTensionChange = now

	return nil
}

// RequiredComponents returns components required by this system
func (fd *Director) RequiredComponents() []ecs.ComponentID {
	// This system doesn't operate on specific components
	return []ecs.ComponentID{}
}

// Update is called every frame
func (fd *Director) Update(deltaTime float64) {
	// Track player
	fd.trackPlayer()

	// Update tension curve
	fd.updateTension(deltaTime)

	// Analyze player behavior (less frequently)
	now := time.Now()
	if now.Sub(fd.lastAnalysisTime).Seconds() >= 5.0 {
		fd.analyzePlayerBehavior()
		fd.lastAnalysisTime = now
	}

	// Check for scare opportunities
	fd.identifyScareOpportunities()

	// Trigger scares if appropriate
	fd.triggerScares()

	// Update active scares
	fd.updateActiveScares(deltaTime)
}

// RecordPlayerAction records a player action for analysis
func (fd *Director) RecordPlayerAction(action PlayerAction) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Add timestamp if not set
	if action.Timestamp.IsZero() {
		action.Timestamp = time.Now()
	}

	// Add action to history
	fd.actionHistory = append(fd.actionHistory, action)

	// Trim history if too long
	if len(fd.actionHistory) > fd.maxHistorySize {
		fd.actionHistory = fd.actionHistory[len(fd.actionHistory)-fd.maxHistorySize:]
	}

	// Update environment awareness from action
	fd.currentAreaType = action.AreaType
	fd.currentLightLevel = action.LightLevel
	fd.currentTimeOfDay = action.TimeOfDay

	// Immediate analysis on certain action types
	if action.EmotionalState >= EmotionFrightened {
		// Player seems scared, note what might have caused it
		fd.analyzeScareResponse(action)
	}
}

// GetTensionLevel returns the current tension level
func (fd *Director) GetTensionLevel() int {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	return fd.tensionLevel
}

// GetTensionName returns the name of the current tension level
func (fd *Director) GetTensionName() string {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	return TensionLevelName[fd.tensionLevel]
}

// GetTensionValue returns the current continuous tension value
func (fd *Director) GetTensionValue() float64 {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	return fd.tensionCurve
}

// GetBehaviorProfile returns the current behavior profile
func (fd *Director) GetBehaviorProfile() *BehaviorProfile {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	return fd.behaviorProfile
}

// GetFearProfile returns the current fear profile
func (fd *Director) GetFearProfile() *FearProfile {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	return fd.fearProfile
}

// SetTensionTarget sets the target tension level
func (fd *Director) SetTensionTarget(target float64) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	fd.targetTension = math.Max(0.0, math.Min(1.0, target))
	fd.tensionDirection = 0 // Reset direction, will be set in updateTension
}

// ForceScare forces a specific type of scare to happen
func (fd *Director) ForceScare(scareType string) bool {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Check if we have a template for this scare type
	template, exists := fd.scareTemplates[scareType]
	if !exists {
		return false
	}

	// Generate a scare from the template
	scare := fd.generateScareFromTemplate(&template, fd.playerPosition)

	// Trigger the scare
	fd.triggerScare(scare)

	return true
}

// SetScareInterval sets the base interval between scares
func (fd *Director) SetScareInterval(seconds float64) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	fd.baseScareInterval = math.Max(10.0, seconds)
}

// SaveProfiles saves the behavior and fear profiles
func (fd *Director) SaveProfiles() error {
	fd.mutex.RLock()
	defer fd.mutex.RUnlock()

	// Ensure directory exists
	err := os.MkdirAll(fd.savePath, os.ModePerm)
	if err != nil {
		return err
	}

	// Save behavior profile
	behaviorPath := filepath.Join(fd.savePath, "behavior_profile.json")
	behaviorData, err := json.MarshalIndent(fd.behaviorProfile, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(behaviorPath, behaviorData, 0644)
	if err != nil {
		return err
	}

	// Save fear profile
	fearPath := filepath.Join(fd.savePath, "fear_profile.json")
	fearData, err := json.MarshalIndent(fd.fearProfile, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fearPath, fearData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// LoadProfiles loads behavior and fear profiles from files
func (fd *Director) LoadProfiles() error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Load behavior profile
	behaviorPath := filepath.Join(fd.savePath, "behavior_profile.json")
	if _, err := os.Stat(behaviorPath); err == nil {
		behaviorData, err := ioutil.ReadFile(behaviorPath)
		if err != nil {
			return err
		}

		var profile BehaviorProfile
		err = json.Unmarshal(behaviorData, &profile)
		if err != nil {
			return err
		}

		fd.behaviorProfile = &profile
	} else {
		return fmt.Errorf("behavior profile not found")
	}

	// Load fear profile
	fearPath := filepath.Join(fd.savePath, "fear_profile.json")
	if _, err := os.Stat(fearPath); err == nil {
		fearData, err := ioutil.ReadFile(fearPath)
		if err != nil {
			return err
		}

		var profile FearProfile
		err = json.Unmarshal(fearData, &profile)
		if err != nil {
			return err
		}

		fd.fearProfile = &profile
	} else {
		return fmt.Errorf("fear profile not found")
	}

	return nil
}

// LoadScareTemplates loads scare templates from files
func (fd *Director) LoadScareTemplates() error {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	templatesPath := filepath.Join(fd.savePath, "scare_templates")
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		return fmt.Errorf("scare templates directory does not exist")
	}

	// Load all template files
	files, err := ioutil.ReadDir(templatesPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(templatesPath, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		var template ScareEvent
		err = json.Unmarshal(data, &template)
		if err != nil {
			continue
		}

		// Add to templates
		fd.scareTemplates[template.Type] = template
	}

	if len(fd.scareTemplates) == 0 {
		return fmt.Errorf("no scare templates found")
	}

	return nil
}

// CreateDefaultScareTemplates creates a set of default scare templates
func (fd *Director) CreateDefaultScareTemplates() error {
	templatesPath := filepath.Join(fd.savePath, "scare_templates")

	// Create directory if it doesn't exist
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		err = os.MkdirAll(templatesPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create default templates
	templates := []ScareEvent{
		// Ambient scares
		{
			ID:                "ambient_sound_1",
			Type:              "ambient_sound",
			Subtype:           "distant",
			Intensity:         0.3,
			Duration:          5.0,
			LightingEffect:    "none",
			SoundEffect:       "distant_howl",
			EntityEffect:      "none",
			EnvironmentEffect: "none",
			EffectRadius:      30.0,
			Cooldown:          30.0,
			Tags:              []string{"ambient", "sound", "subtle"},
		},
		{
			ID:                "ambient_sound_2",
			Type:              "ambient_sound",
			Subtype:           "close",
			Intensity:         0.5,
			Duration:          3.0,
			LightingEffect:    "none",
			SoundEffect:       "nearby_breaking",
			EntityEffect:      "none",
			EnvironmentEffect: "none",
			EffectRadius:      10.0,
			Cooldown:          45.0,
			Tags:              []string{"ambient", "sound", "moderate"},
		},
		{
			ID:                "ambient_visual_1",
			Type:              "ambient_visual",
			Subtype:           "shadow",
			Intensity:         0.4,
			Duration:          2.0,
			LightingEffect:    "shadow_movement",
			SoundEffect:       "none",
			EntityEffect:      "none",
			EnvironmentEffect: "none",
			EffectRadius:      15.0,
			Cooldown:          60.0,
			Tags:              []string{"ambient", "visual", "subtle"},
		},

		// Environment scares
		{
			ID:                "environment_1",
			Type:              "environment",
			Subtype:           "weather",
			Intensity:         0.4,
			Duration:          20.0,
			LightingEffect:    "darkening",
			SoundEffect:       "thunder",
			EntityEffect:      "none",
			EnvironmentEffect: "rain_intensify",
			EffectRadius:      100.0,
			Cooldown:          180.0,
			Tags:              []string{"environment", "weather", "moderate"},
		},
		{
			ID:                "environment_2",
			Type:              "environment",
			Subtype:           "object",
			Intensity:         0.5,
			Duration:          5.0,
			LightingEffect:    "none",
			SoundEffect:       "breaking",
			EntityEffect:      "none",
			EnvironmentEffect: "object_break",
			EffectRadius:      8.0,
			Cooldown:          90.0,
			Tags:              []string{"environment", "object", "moderate"},
		},

		// Entity scares
		{
			ID:                "entity_1",
			Type:              "entity",
			Subtype:           "creature",
			Intensity:         0.6,
			Duration:          15.0,
			LightingEffect:    "none",
			SoundEffect:       "growl",
			EntityEffect:      "creature_appear",
			EnvironmentEffect: "none",
			EffectRadius:      20.0,
			Cooldown:          300.0,
			Tags:              []string{"entity", "creature", "intense"},
		},
		{
			ID:                "entity_2",
			Type:              "entity",
			Subtype:           "stalker",
			Intensity:         0.7,
			Duration:          30.0,
			LightingEffect:    "none",
			SoundEffect:       "footsteps",
			EntityEffect:      "stalker",
			EnvironmentEffect: "none",
			EffectRadius:      25.0,
			Cooldown:          360.0,
			Tags:              []string{"entity", "stalker", "intense"},
		},

		// Jump scares
		{
			ID:                "jumpscare_1",
			Type:              "jumpscare",
			Subtype:           "visual",
			Intensity:         0.8,
			Duration:          1.0,
			LightingEffect:    "flash",
			SoundEffect:       "scare_sound",
			EntityEffect:      "jump_visual",
			EnvironmentEffect: "none",
			EffectRadius:      5.0,
			Cooldown:          600.0,
			Tags:              []string{"jumpscare", "visual", "intense"},
		},

		// Psychological scares
		{
			ID:                "psychological_1",
			Type:              "psychological",
			Subtype:           "paranoia",
			Intensity:         0.5,
			Duration:          45.0,
			LightingEffect:    "subtle_pulsing",
			SoundEffect:       "whispers",
			EntityEffect:      "none",
			EnvironmentEffect: "subtle_movement",
			EffectRadius:      15.0,
			Cooldown:          300.0,
			Tags:              []string{"psychological", "paranoia", "moderate"},
		},

		// Metamorphosis scares
		{
			ID:                "metamorph_1",
			Type:              "metamorphosis",
			Subtype:           "environment",
			Intensity:         0.7,
			Duration:          60.0,
			LightingEffect:    "color_shift",
			SoundEffect:       "reality_shift",
			EntityEffect:      "none",
			EnvironmentEffect: "metamorph_environment",
			EffectRadius:      30.0,
			Cooldown:          900.0,
			Tags:              []string{"metamorphosis", "environment", "strong"},
		},
	}

	// Save each template to a file
	for _, template := range templates {
		filePath := filepath.Join(templatesPath, template.ID+".json")
		data, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			continue
		}

		err = ioutil.WriteFile(filePath, data, 0644)
		if err != nil {
			continue
		}
	}

	return nil
}

// trackPlayer finds and tracks the player's entity
func (fd *Director) trackPlayer() {
	// Find player entity if not already tracked
	if fd.playerID == "" {
		playerEntities := fd.world.GetEntitiesWithTag("player")
		if len(playerEntities) > 0 {
			fd.playerID = playerEntities[0].ID
		} else {
			return
		}
	}

	// Get player entity
	player, exists := fd.world.GetEntity(fd.playerID)
	if !exists {
		// Player entity no longer exists, reset tracking
		fd.playerID = ""
		return
	}

	// Get player position
	transformComp, has := player.GetComponent(ecs.TransformComponentID)
	if !has {
		return
	}

	transform := transformComp.(*ecs.TransformComponent)
	fd.playerPosition = transform.Position
	fd.playerLastSeen = time.Now()
}

// updateTension updates the tension curve
func (fd *Director) updateTension(deltaTime float64) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Determine tension direction
	if fd.tensionCurve < fd.targetTension {
		fd.tensionDirection = 1 // Increasing
	} else if fd.tensionCurve > fd.targetTension {
		fd.tensionDirection = -1 // Decreasing
	} else {
		fd.tensionDirection = 0 // Stable
	}

	// Update tension based on direction
	if fd.tensionDirection > 0 {
		// Increasing tension
		fd.tensionCurve += fd.tensionChangeRate * deltaTime
		if fd.tensionCurve > fd.targetTension {
			fd.tensionCurve = fd.targetTension
		}
	} else if fd.tensionDirection < 0 {
		// Decreasing tension
		fd.tensionCurve -= fd.tensionChangeRate * deltaTime
		if fd.tensionCurve < fd.targetTension {
			fd.tensionCurve = fd.targetTension
		}
	}

	// Clamp tension to 0-1 range
	fd.tensionCurve = math.Max(0.0, math.Min(1.0, fd.tensionCurve))

	// Determine new tension level (0-4)
	newLevel := int(fd.tensionCurve * 4)

	// Check if tension level changed
	if newLevel != fd.tensionLevel {
		// Record when the level changed
		fd.lastTensionChange = time.Now()
		fd.tensionLevel = newLevel

		// Update tension phase
		if newLevel >= 3 {
			fd.tensionPhase = "peak"
		} else if fd.tensionDirection > 0 {
			fd.tensionPhase = "build"
		} else if fd.tensionDirection < 0 {
			fd.tensionPhase = "release"
		} else {
			fd.tensionPhase = "calm"
		}
	}

	// Check if we need to start decreasing tension after peaking too long
	if fd.tensionPhase == "peak" {
		peakDuration := time.Since(fd.lastTensionChange).Seconds()
		if peakDuration > fd.maxTensionTime {
			// We've been at peak tension too long, start decreasing
			fd.targetTension = 0.3 // Target a moderate-low tension
		}
	}
}

// analyzePlayerBehavior performs analysis on recorded player actions
func (fd *Director) analyzePlayerBehavior() {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Skip if not enough data
	if len(fd.actionHistory) < 5 {
		return
	}

	// Get recent actions (last 10)
	recentActions := fd.actionHistory
	if len(recentActions) > 10 {
		recentActions = recentActions[len(recentActions)-10:]
	}

	// Count action types
	actionCounts := map[ActionType]int{}
	for _, action := range recentActions {
		actionCounts[action.Type]++
	}

	// Calculate behavior metrics

	// Calculate risk aversion
	// Higher values of running, hiding indicate higher risk aversion
	runningCount := actionCounts[ActionRunning]
	hidingCount := actionCounts[ActionHiding]
	fightingCount := actionCounts[ActionFighting]

	totalActions := len(recentActions)
	riskAversionScore := float64(runningCount+hidingCount) / float64(totalActions)
	aggressionScore := float64(fightingCount) / float64(totalActions)

	// Apply exponential smoothing to behavior profile (alpha = 0.3)
	alpha := 0.3
	fd.behaviorProfile.RiskAversion = alpha*riskAversionScore + (1-alpha)*fd.behaviorProfile.RiskAversion
	fd.behaviorProfile.Aggression = alpha*aggressionScore + (1-alpha)*fd.behaviorProfile.Aggression

	// Calculate thoroughness from exploring and inspecting actions
	exploringCount := actionCounts[ActionExploring]
	inspectingCount := actionCounts[ActionInspecting]

	thoroughnessScore := float64(exploringCount+inspectingCount) / float64(totalActions)
	fd.behaviorProfile.Thoroughness = alpha*thoroughnessScore + (1-alpha)*fd.behaviorProfile.Thoroughness

	// Calculate patience from resting and crafting actions
	restingCount := actionCounts[ActionResting]
	craftingCount := actionCounts[ActionCrafting]

	patienceScore := float64(restingCount+craftingCount) / float64(totalActions)
	fd.behaviorProfile.Patience = alpha*patienceScore + (1-alpha)*fd.behaviorProfile.Patience

	// Analyze looking behavior
	lookingAroundCount := 0
	for _, action := range recentActions {
		if action.LookingAround {
			lookingAroundCount++
		}
	}

	cautionScore := float64(lookingAroundCount) / float64(totalActions)
	fd.behaviorProfile.Caution = alpha*cautionScore + (1-alpha)*fd.behaviorProfile.Caution

	// Update movement patterns
	if len(recentActions) >= 3 {
		// Check if player tends to return to same areas
		// by analyzing position history
		// This is a simple implementation - could be more sophisticated
		revisitScore := 0.0
		uniqueAreas := map[string]bool{}

		for _, action := range recentActions {
			// Round position to create area identifier
			areaX := math.Floor(action.Position.X/10.0) * 10.0
			areaZ := math.Floor(action.Position.Z/10.0) * 10.0
			areaID := fmt.Sprintf("%.0f_%.0f", areaX, areaZ)

			uniqueAreas[areaID] = true
		}

		revisitScore = 1.0 - float64(len(uniqueAreas))/float64(len(recentActions))

		// Update movement patterns map
		if _, exists := fd.behaviorProfile.MovementPatterns["revisits"]; !exists {
			fd.behaviorProfile.MovementPatterns["revisits"] = 0.5 // Initial middle value
		}

		fd.behaviorProfile.MovementPatterns["revisits"] = alpha*revisitScore +
			(1-alpha)*fd.behaviorProfile.MovementPatterns["revisits"]
	}

	// Update comfort zones based on where player spends time
	// Calculate time spent in each area type
	areaTimeCounts := map[string]int{}
	for _, action := range recentActions {
		if action.AreaType != "" {
			areaTimeCounts[action.AreaType]++
		}
	}

	// Update comfort zones
	for area, count := range areaTimeCounts {
		score := float64(count) / float64(totalActions)

		if _, exists := fd.behaviorProfile.ComfortZones[area]; !exists {
			fd.behaviorProfile.ComfortZones[area] = 0.5 // Initial middle value
		}

		fd.behaviorProfile.ComfortZones[area] = alpha*score +
			(1-alpha)*fd.behaviorProfile.ComfortZones[area]
	}

	// Update fear triggers based on emotional responses
	for _, action := range recentActions {
		if action.EmotionalState >= EmotionNervous {
			// This action made the player nervous or scared
			// Look at context tags
			for _, tag := range action.ContextTags {
				if _, exists := fd.fearProfile.ContextualFears[tag]; !exists {
					fd.fearProfile.ContextualFears[tag] = 0.5 // Initial middle value
				}

				// Increase fear rating for this tag
				// Higher emotional state means bigger increase
				increase := float64(action.EmotionalState) * 0.1
				fd.fearProfile.ContextualFears[tag] = math.Min(1.0,
					fd.fearProfile.ContextualFears[tag]+increase)
			}

			// Update environment fears
			if action.AreaType != "" {
				if _, exists := fd.fearProfile.EnvironmentFears[action.AreaType]; !exists {
					fd.fearProfile.EnvironmentFears[action.AreaType] = 0.5
				}

				increase := float64(action.EmotionalState) * 0.05
				fd.fearProfile.EnvironmentFears[action.AreaType] = math.Min(1.0,
					fd.fearProfile.EnvironmentFears[action.AreaType]+increase)
			}

			// Update entity fears if a target is specified
			if action.Target != "" {
				targetStr := string(action.Target)
				if _, exists := fd.fearProfile.EntityFears[targetStr]; !exists {
					fd.fearProfile.EntityFears[targetStr] = 0.5
				}

				increase := float64(action.EmotionalState) * 0.15
				fd.fearProfile.EntityFears[targetStr] = math.Min(1.0,
					fd.fearProfile.EntityFears[targetStr]+increase)
			}

			// Update specific fears based on context
			if action.LightLevel < 0.3 {
				// Dark situation
				fd.fearProfile.DarknessFear = math.Min(1.0,
					fd.fearProfile.DarknessFear+float64(action.EmotionalState)*0.05)
			}
		}
	}
}

// analyzeScareResponse analyzes the player's response to a scare
func (fd *Director) analyzeScareResponse(action PlayerAction) {
	// Look for recent scares that might have caused this response
	recentScareTime := 10.0 // Look for scares in last 10 seconds

	// Find scares that happened recently
	now := time.Now()
	recentScares := []*ScareEvent{}

	for _, scare := range fd.currentScares {
		// Check if scare is recent enough
		if now.Sub(scare.SuccessRating).Seconds() <= recentScareTime {
			recentScares = append(recentScares, scare)
		}
	}

	if len(recentScares) == 0 {
		// No recent scares, look in history
		for i := len(fd.scareHistory) - 1; i >= 0; i-- {
			scare := fd.scareHistory[i]
			if now.Sub(scare.SuccessRating).Seconds() <= recentScareTime {
				recentScares = append(recentScares, &scare)
			} else {
				// Too old, stop looking
				break
			}
		}
	}

	if len(recentScares) == 0 {
		// No recent scares found
		return
	}

	// Sort scares by recency (most recent first)
	sort.Slice(recentScares, func(i, j int) bool {
		iTime := recentScares[i].SuccessRating
		jTime := recentScares[j].SuccessRating
		return iTime.After(jTime)
	})

	// Analyze most recent scare
	mostRecentScare := recentScares[0]

	// Update effective scares mapping
	if _, exists := fd.fearProfile.EffectiveScares[mostRecentScare.Type]; !exists {
		fd.fearProfile.EffectiveScares[mostRecentScare.Type] = 0.5
	}

	// Calculate effectiveness based on emotional response
	effectiveness := float64(action.EmotionalState) / 4.0

	// Update effective scares
	fd.fearProfile.EffectiveScares[mostRecentScare.Type] = 0.7*effectiveness +
		0.3*fd.fearProfile.EffectiveScares[mostRecentScare.Type]

	// Update successful scares counter
	fd.successfulScares[mostRecentScare.Type]++
}

// identifyScareOpportunities looks for good opportunities to scare the player
func (fd *Director) identifyScareOpportunities() {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Skip if no player
	if fd.playerID == "" {
		return
	}

	// Skip if we've recently triggered a scare
	timeSinceLastScare := time.Since(fd.lastScareTime).Seconds()
	if timeSinceLastScare < fd.minScareInterval {
		return
	}

	// Clear old opportunities
	fd.scareOpportunities = []ScareOpportunity{}

	// Skip opportunity generation if tension is rising but not yet high enough
	if fd.tensionPhase == "build" && fd.tensionLevel < 2 {
		return
	}

	// Get player state from most recent action
	var playerState string
	if len(fd.actionHistory) > 0 {
		playerState = string(fd.actionHistory[len(fd.actionHistory)-1].Type)
	} else {
		playerState = string(ActionNone)
	}

	// Generate new opportunities based on player state and context
	switch playerState {
	case string(ActionMoving), string(ActionExploring):
		// Good time for ambient scares or environment scares
		fd.addAmbientScareOpportunity()

		// If tension is high, opportunity for entity scares
		if fd.tensionLevel >= 3 {
			fd.addEntityScareOpportunity()
		}

	case string(ActionRunning):
		// Player already running - good for chase enhancement
		if fd.tensionLevel >= 2 {
			fd.addChaseScare()
		}

	case string(ActionHiding):
		// Player hiding - good for tension or "false safety" scares
		fd.addHidingScareOpportunity()

	case string(ActionInspecting):
		// Player inspecting something - good for focus break
		fd.addInspectionScareOpportunity()

	case string(ActionResting):
		// Player resting - good time for subtle buildup
		fd.addRestingScareOpportunity()
	}

	// Add metamorphosis opportunity if tension is very high
	if fd.tensionLevel >= 3 && fd.tensionPhase == "peak" {
		fd.addMetamorphosisOpportunity()
	}
}

// triggerScares triggers scare events when appropriate
func (fd *Director) triggerScares() {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	// Skip if no opportunities
	if len(fd.scareOpportunities) == 0 {
		return
	}

	// Sort opportunities by estimated value (most effective first)
	sort.Slice(fd.scareOpportunities, func(i, j int) bool {
		return fd.scareOpportunities[i].EstimatedValue > fd.scareOpportunities[j].EstimatedValue
	})

	// Get best opportunity
	bestOpportunity := fd.scareOpportunities[0]

	// Skip if we need to wait longer
	if bestOpportunity.OptimalTiming > 0 &&
		time.Since(bestOpportunity.Timestamp).Seconds() < bestOpportunity.OptimalTiming {
		return
	}

	// Skip if estimated value is too low
	if bestOpportunity.EstimatedValue < 0.3 {
		return
	}

	// Generate a scare based on opportunity
	if len(bestOpportunity.ScareTypes) > 0 {
		scareType := bestOpportunity.ScareTypes[0]

		// If we have multiple types, pick the most effective one
		if len(bestOpportunity.ScareTypes) > 1 {
			bestScareType := ""
			bestEffectiveness := 0.0

			for _, st := range bestOpportunity.ScareTypes {
				if eff, exists := fd.fearProfile.EffectiveScares[st]; exists {
					if eff > bestEffectiveness {
						bestEffectiveness = eff
						bestScareType = st
					}
				}
			}

			if bestScareType != "" {
				scareType = bestScareType
			}
		}

		// Check if we have a template for this scare type
		template, exists := fd.scareTemplates[scareType]
		if exists {
			// Check cooldown
			if cooldownTime, hasCooldown := fd.scareCooldowns[scareType]; hasCooldown {
				if time.Now().Before(cooldownTime) {
					// Scare is on cooldown, try next opportunity
					if len(fd.scareOpportunities) > 1 {
						fd.scareOpportunities = fd.scareOpportunities[1:]
						fd.triggerScares() // Recursively try next opportunity
					}
					return
				}
			}

			// Generate scare from template
			scare := fd.generateScareFromTemplate(&template, bestOpportunity.Position)

			// Trigger the scare
			fd.triggerScare(scare)

			// Remove this opportunity
			fd.scareOpportunities = fd.scareOpportunities[1:]
		}
	}
}

// updateActiveScares updates all currently active scares
func (fd *Director) updateActiveScares(deltaTime float64) {
	fd.mutex.Lock()
	defer fd.mutex.Unlock()

	now := time.Now()

	// Check each active scare
	for id, scare := range fd.currentScares {
		// Check if scare has expired
		if now.Sub(scare.SuccessRating).Seconds() >= scare.Duration {
			// Scare is over, remove from active scares
			delete(fd.currentScares, id)

			// Add to history
			fd.scareHistory = append(fd.scareHistory, *scare)

			// Trim history if too long
			if len(fd.scareHistory) > 50 {
				fd.scareHistory = fd.scareHistory[len(fd.scareHistory)-50:]
			}
		}
	}
}

// generateScareFromTemplate creates a scare event from a template
func (fd *Director) generateScareFromTemplate(template *ScareEvent, position ecs.Vector3) *ScareEvent {
	// Create a copy of the template
	scare := *template

	// Generate unique ID
	scare.ID = fmt.Sprintf("%s_%d", template.Type, time.Now().UnixNano())

	// Set positions
	scare.StartPosition = position

	// Calculate target position (usually in front of player)
	// Get player entity
	player, exists := fd.world.GetEntity(fd.playerID)
	if exists {
		// Get player transform
		transformComp, has := player.GetComponent(ecs.TransformComponentID)
		if has {
			transform := transformComp.(*ecs.TransformComponent)

			// Get forward direction
			forward := transform.Forward()

			// Set target position in front of player
			scare.TargetPosition = position.Add(forward.Multiply(template.EffectRadius * 0.7))
		} else {
			// Default target is same as start position
			scare.TargetPosition = position
		}
	} else {
		// Default target is same as start position
		scare.TargetPosition = position
	}

	// Mark creation time as now
	scare.SuccessRating = time.Now()

	return &scare
}

// triggerScare triggers a specific scare event
func (fd *Director) triggerScare(scare *ScareEvent) {
	// Add to current scares
	fd.currentScares[scare.ID] = scare

	// Set cooldown
	fd.scareCooldowns[scare.Type] = time.Now().Add(time.Duration(scare.Cooldown * float64(time.Second)))

	// Update last scare time
	fd.lastScareTime = time.Now()

	// Update tension target based on intensity
	newTarget := fd.tensionCurve + scare.Intensity*0.3
	fd.targetTension = math.Min(1.0, newTarget)

	// Invoke callback if set
	if fd.OnScareTriggered != nil {
		fd.OnScareTriggered(*scare)
	}
}

// Helper methods for generating scare opportunities

// addAmbientScareOpportunity adds an ambient scare opportunity
func (fd *Director) addAmbientScareOpportunity() {
	// Base value depends on tension level
	value := 0.3 + float64(fd.tensionLevel)*0.1

	// Adjust based on area type
	if fd.currentAreaType == "forest" && fd.currentTimeOfDay < 0.25 {
		// Dark forest is scarier
		value += 0.1
	} else if fd.currentAreaType == "cave" {
		// Caves are naturally tense
		value += 0.15
	}

	// Adjust based on light level
	if fd.currentLightLevel < 0.3 {
		// Darker is scarier for ambient scares
		value += (0.3 - fd.currentLightLevel) * 0.5
	}

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"ambient_sound", "ambient_visual"},
		Position:       fd.playerPosition,
		OptimalTiming:  2.0 + rand.Float64()*5.0, // 2-7 seconds delay
		EstimatedValue: value,
		PlayerState:    "moving",
		Context:        fd.currentAreaType,
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addEntityScareOpportunity adds an entity-based scare opportunity
func (fd *Director) addEntityScareOpportunity() {
	// Entity scares are more impactful
	value := 0.5 + float64(fd.tensionLevel)*0.1

	// Adjust based on behavior profile
	if fd.behaviorProfile.RiskAversion > 0.6 {
		// Risk averse players are more scared by entities
		value += 0.2
	}

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"entity"},
		Position:       fd.playerPosition,
		OptimalTiming:  1.0 + rand.Float64()*3.0, // 1-4 seconds delay
		EstimatedValue: value,
		PlayerState:    "exploring",
		Context:        fd.currentAreaType,
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addChaseScare adds a scare opportunity during chase
func (fd *Director) addChaseScare() {
	// Chase enhancement is very effective at high tension
	value := 0.6 + float64(fd.tensionLevel)*0.1

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"entity", "environment"},
		Position:       fd.playerPosition,
		OptimalTiming:  0.5 + rand.Float64()*1.5, // Quick response, 0.5-2 seconds
		EstimatedValue: value,
		PlayerState:    "running",
		Context:        "chase",
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addHidingScareOpportunity adds a scare for when player is hiding
func (fd *Director) addHidingScareOpportunity() {
	// Base value
	value := 0.4 + float64(fd.tensionLevel)*0.1

	// Higher value if tension is already high
	if fd.tensionLevel >= 3 {
		value += 0.2
	}

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"ambient_sound", "psychological"},
		Position:       fd.playerPosition,
		OptimalTiming:  3.0 + rand.Float64()*5.0, // 3-8 seconds (let them think they're safe)
		EstimatedValue: value,
		PlayerState:    "hiding",
		Context:        "false_safety",
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addInspectionScareOpportunity adds a scare when player is inspecting something
func (fd *Director) addInspectionScareOpportunity() {
	// Good for jump scares or sudden events
	value := 0.5 + float64(fd.tensionLevel)*0.1

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"jumpscare", "environment"},
		Position:       fd.playerPosition,
		OptimalTiming:  1.0 + rand.Float64()*2.0, // 1-3 seconds
		EstimatedValue: value,
		PlayerState:    "inspecting",
		Context:        "focus_break",
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addRestingScareOpportunity adds a scare when player is resting
func (fd *Director) addRestingScareOpportunity() {
	// Subtle build-up
	value := 0.3 + float64(fd.tensionLevel)*0.1

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"ambient_sound", "ambient_visual", "psychological"},
		Position:       fd.playerPosition,
		OptimalTiming:  5.0 + rand.Float64()*10.0, // 5-15 seconds (slow build)
		EstimatedValue: value,
		PlayerState:    "resting",
		Context:        "calm_before_storm",
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// addMetamorphosisOpportunity adds an opportunity for a metamorphosis scare
func (fd *Director) addMetamorphosisOpportunity() {
	// Metamorphosis scares are rare and powerful
	value := 0.7 + float64(fd.tensionLevel)*0.1

	// Adjust based on current phase
	value = math.Min(0.95, value)

	// Create opportunity
	opportunity := ScareOpportunity{
		Timestamp:      time.Now(),
		ScareTypes:     []string{"metamorphosis"},
		Position:       fd.playerPosition,
		OptimalTiming:  2.0 + rand.Float64()*3.0, // 2-5 seconds
		EstimatedValue: value,
		PlayerState:    "any",
		Context:        "reality_shift",
		TensionLevel:   fd.tensionLevel,
	}

	fd.scareOpportunities = append(fd.scareOpportunities, opportunity)
}

// NewDefaultBehaviorProfile creates a default behavior profile
func NewDefaultBehaviorProfile() *BehaviorProfile {
	return &BehaviorProfile{
		RiskAversion:     0.5,
		Thoroughness:     0.5,
		Caution:          0.5,
		Aggression:       0.5,
		Patience:         0.5,
		Curiosity:        0.6, // Slightly elevated curiosity by default
		FearTriggers:     make(map[string]float64),
		ComfortZones:     make(map[string]float64),
		MovementPatterns: make(map[string]float64),
		DecisionPatterns: make(map[string]float64),
		TimePatterns:     make(map[string]float64),
		ScareResponses:   make(map[string]float64),
		Adaptability:     0.5,
		Predictability:   0.5,
	}
}

// NewDefaultFearProfile creates a default fear profile
func NewDefaultFearProfile() *FearProfile {
	return &FearProfile{
		DarknessFear:       0.7, // Most people fear darkness to some degree
		EnclosedSpacesFear: 0.5,
		HeightsFear:        0.4,
		WaterFear:          0.3,
		FireFear:           0.4,
		MonstersFear:       0.6,
		InsectsFear:        0.5,
		GoreFear:           0.5,
		JumpscaresFear:     0.8, // Jumpscares are effective on most people
		PsychologicalFear:  0.6,
		NoiseFear:          0.6,
		SilenceFear:        0.4,
		UnknownFear:        0.7, // Fear of the unknown is common
		IsolationFear:      0.5,
		ParanormalFear:     0.6,
		EntityFears:        make(map[string]float64),
		EnvironmentFears:   make(map[string]float64),
		ContextualFears:    make(map[string]float64),
		EffectiveScares:    make(map[string]float64),
		Habituation:        0.3, // Low value means slower habituation
		Recovery:           0.5,
	}
}
