# Phase 1: Core Functionality Implementation Plan
**Duration: 2-3 weeks**  
**Goal: Establish solid foundation with enhanced CLI and core missing features**

## Overview
Phase 1 focuses on building the essential functionality that users need immediately while establishing the foundation for future enhancements. This phase prioritizes user-facing features and basic API improvements.

---

## Subphase 1A: Package Reorganization & CLI Foundation (Days 1-3)

### Objectives
- Restructure packages for better maintainability
- Set up cobra-based CLI framework
- Establish consistent naming conventions

### Tasks

#### 1A.1: Package Structure Refactoring
**Duration: 1 day**

```
Current Structure → New Structure
garth/              pkg/garmin/
├── client/         ├── client.go      # Main client interface
├── data/           ├── activities.go  # Activity operations
├── stats/          ├── health.go      # Health data operations
├── sso/            ├── stats.go       # Statistics operations
├── oauth/          ├── auth.go        # Authentication
└── ...             └── types.go       # Public types

                    internal/
                    ├── api/           # Low-level API client
                    ├── auth/          # Auth implementation
                    ├── data/          # Data processing
                    └── utils/         # Internal utilities

                    cmd/garth/
                    ├── main.go        # CLI entry point
                    ├── root.go        # Root command
                    ├── auth.go        # Auth commands
                    ├── activities.go  # Activity commands
                    ├── health.go      # Health commands
                    └── stats.go       # Stats commands
```

**Deliverables:**
- [ ] New package structure implemented
- [ ] All imports updated
- [ ] No breaking changes to existing functionality
- [ ] Package documentation updated

#### 1A.2: CLI Framework Setup
**Duration: 1 day**

```go
// cmd/garth/root.go
var rootCmd = &cobra.Command{
    Use:   "garth",
    Short: "Garmin Connect CLI tool",
    Long:  `A comprehensive CLI tool for interacting with Garmin Connect`,
}

// Global flags
var (
    configFile string
    outputFormat string // json, table, csv
    verbose bool
    dateFrom string
    dateTo string
)
```

**Tasks:**
- [ ] Install and configure cobra
- [ ] Create root command with global flags
- [ ] Implement configuration file loading
- [ ] Add output formatting infrastructure
- [ ] Create help text and usage examples

**Deliverables:**
- [ ] Working CLI framework with `garth --help`
- [ ] Configuration file support
- [ ] Output formatting (JSON, table, CSV)

#### 1A.3: Configuration Management
**Duration: 1 day**

```go
// internal/config/config.go
type Config struct {
    Auth struct {
        Email    string `yaml:"email"`
        Domain   string `yaml:"domain"`
        Session  string `yaml:"session_file"`
    } `yaml:"auth"`
    
    Output struct {
        Format string `yaml:"format"`
        File   string `yaml:"file"`
    } `yaml:"output"`
    
    Cache struct {
        Enabled bool   `yaml:"enabled"`
        TTL     string `yaml:"ttl"`
        Dir     string `yaml:"dir"`
    } `yaml:"cache"`
}
```

**Tasks:**
- [ ] Design configuration schema
- [ ] Implement config file loading/saving
- [ ] Add environment variable support
- [ ] Create config validation
- [ ] Add config commands (`garth config init`, `garth config show`)

**Deliverables:**
- [ ] Configuration system working
- [ ] Default config file created
- [ ] Config commands implemented

---

## Subphase 1B: Enhanced CLI Commands (Days 4-7)

### Objectives
- Implement all major CLI commands
- Add interactive features
- Ensure consistent user experience

### Tasks

#### 1B.1: Authentication Commands
**Duration: 1 day**

```bash
# Target CLI interface
garth auth login                    # Interactive login
garth auth login --email user@example.com --password-stdin
garth auth logout                   # Clear session
garth auth status                   # Show auth status
garth auth refresh                  # Refresh tokens
```

```go
// cmd/garth/auth.go
var authCmd = &cobra.Command{
    Use:   "auth",
    Short: "Authentication management",
}

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Login to Garmin Connect",
    RunE:  runLogin,
}
```

**Tasks:**
- [x] Implement `auth login` with interactive prompts
- [x] Add `auth logout` functionality
- [x] Create `auth status` command
- [x] Implement secure password input
- [ ] Add MFA support (prepare for future)
- [x] Session validation and refresh

**Deliverables:**
- [x] All auth commands working
- [x] Secure credential handling
- [x] Session persistence working

#### 1B.2: Activity Commands
**Duration: 2 days**

```bash
# Target CLI interface
garth activities list                           # Recent activities
garth activities list --limit 50 --type running
garth activities get 12345678                   # Activity details  
garth activities download 12345678 --format gpx
garth activities search --query "morning run"
```

```go
// pkg/garmin/activities.go
type ActivityOptions struct {
    Limit      int
    Offset     int
    ActivityType string
    DateFrom   time.Time
    DateTo     time.Time
}

type ActivityDetail struct {
    BasicInfo Activity
    Summary   ActivitySummary
    Laps      []Lap
    Metrics   []Metric
}
```

**Tasks:**
- [x] Enhanced activity listing with filters
- [x] Activity detail fetching
- [x] Search functionality
- [x] Table formatting for activity lists
- [x] Activity download preparation (basic structure)
- [x] Date range filtering
- [x] Activity type filtering

**Deliverables:**
- [x] `activities list` with all filtering options
- [x] `activities get` showing detailed info
- [x] `activities search` functionality
- [x] Proper error handling and user feedback

#### 1B.3: Health Data Commands  
**Duration: 2 days**

```bash
# Target CLI interface
garth health sleep --from 2024-01-01 --to 2024-01-07
garth health hrv --days 30
garth health stress --week
garth health bodybattery --yesterday
```

**Tasks:**
- [x] Implement all health data commands
- [x] Add date range parsing utilities
- [x] Create consistent output formatting
- [x] Add data aggregation options
- [ ] Implement caching for expensive operations
- [x] Error handling for missing data

**Deliverables:**
- [x] All health commands working
- [x] Consistent date filtering across commands
- [x] Proper data formatting and display

#### 1B.4: Statistics Commands
**Duration: 1 day**

```bash
# Target CLI interface  
garth stats steps --month
garth stats distance --year
garth stats calories --from 2024-01-01
```

**Tasks:**
- [x] Implement statistics commands
- [x] Add aggregation periods (day, week, month, year)
- [x] Create summary statistics
- [ ] Add trend analysis
- [x] Implement data export options

**Deliverables:**
- [x] All stats commands working
- [x] Multiple aggregation options
- [x] Export functionality

---

## Subphase 1C: Activity Download Implementation (Days 8-12)

### Objectives
- Implement activity file downloading
- Support multiple formats (GPX, TCX, FIT)
- Add batch download capabilities

### Tasks

#### 1C.1: Core Download Infrastructure
**Duration: 2 days**

```go
// pkg/garmin/activities.go
type DownloadOptions struct {
    Format     string // "gpx", "tcx", "fit", "csv"
    Original   bool   // Download original uploaded file
    OutputDir  string
    Filename   string
}

func (c *Client) DownloadActivity(id string, opts *DownloadOptions) error {
    // Implementation
}
```

**Tasks:**
- [x] Research Garmin's download endpoints
- [x] Implement format detection and conversion
- [x] Add file writing with proper naming
- [x] Implement progress indication
- [x] Add download validation
- [x] Error handling for failed downloads

**Deliverables:**
- [x] Working download for at least GPX format
- [x] Progress indication during download
- [x] Proper error handling

#### 1C.2: Multi-Format Support
**Duration: 2 days**

**Tasks:**
- [x] Implement TCX format download
- [x] Implement FIT format download (if available)
- [x] Add CSV export for activity summaries
- [x] Format validation and conversion
- [x] Add format-specific options

**Deliverables:**
- [x] Support for GPX, TCX, and CSV formats
- [x] Format auto-detection
- [x] Format-specific download options

#### 1C.3: Batch Download Features
**Duration: 1 day**

```bash
# Target functionality
garth activities download --all --type running --format gpx
garth activities download --from 2024-01-01 --to 2024-01-31
```

**Tasks:**
- [x] Implement batch download with filtering
- [x] Add parallel download support
- [x] Progress bars for multiple downloads
- [ ] Resume interrupted downloads
- [x] Duplicate detection and handling

**Deliverables:**
- [x] Batch download working
- [x] Parallel processing implemented
- [ ] Resume capability

---

## Subphase 1D: Missing Health Data Types (Days 13-15)

### Objectives
- Implement VO2 max data fetching
- Add heart rate zones
- Complete missing health metrics

### Tasks

#### 1D.1: VO2 Max Implementation
**Duration: 1 day**

```go
// pkg/garmin/health.go
type VO2MaxData struct {
    Running *VO2MaxReading `json:"running"`
    Cycling *VO2MaxReading `json:"cycling"`
    Updated time.Time      `json:"updated"`
    History []VO2MaxHistory `json:"history"`
}

type VO2MaxReading struct {
    Value      float64   `json:"value"`
    UpdatedAt  time.Time `json:"updated_at"`
    Source     string    `json:"source"`
    Confidence string    `json:"confidence"`
}
```

**Tasks:**
- [x] Research VO2 max API endpoints
- [x] Implement data fetching
- [x] Add historical data support
- [x] Create CLI command
- [x] Add data validation
- [x] Format output appropriately

**Deliverables:**
- [x] `garth health vo2max` command working
- [x] Historical data support
- [x] Both running and cycling metrics

#### 1D.2: Heart Rate Zones
**Duration: 1 day**

```go
type HeartRateZones struct {
    RestingHR    int      `json:"resting_hr"`
    MaxHR        int      `json:"max_hr"`
    LactateThreshold int  `json:"lactate_threshold"`
    Zones        []HRZone `json:"zones"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type HRZone struct {
    Zone     int `json:"zone"`
    MinBPM   int `json:"min_bpm"`
    MaxBPM   int `json:"max_bpm"`
    Name     string `json:"name"`
}
```

**Tasks:**
- [x] Implement HR zones API calls
- [x] Add zone calculation logic
- [x] Create CLI command
- [x] Add zone analysis features
- [x] Implement zone updates (if possible)

**Deliverables:**
- [x] `garth health hr-zones` command
- [x] Zone calculation and display
- [ ] Integration with other health metrics

#### 1D.3: Additional Health Metrics
**Duration: 1 day**

```go
type WellnessData struct {
    Date         time.Time `json:"date"`
    RestingHR    *int      `json:"resting_hr"`
    Weight       *float64  `json:"weight"`
    BodyFat      *float64  `json:"body_fat"`
    BMI          *float64  `json:"bmi"`
    BodyWater    *float64  `json:"body_water"`
    BoneMass     *float64  `json:"bone_mass"`
    MuscleMass   *float64  `json:"muscle_mass"`
}
```

**Tasks:**
- [ ] Research additional wellness endpoints
- [ ] Implement body composition data
- [ ] Add resting heart rate trends
- [ ] Create comprehensive wellness command
- [ ] Add data correlation features

**Deliverables:**
- [ ] Additional health metrics available
- [ ] Wellness overview command
- [ ] Data trend analysis

---

## Phase 1 Testing & Quality Assurance (Days 14-15)

### Tasks

#### Integration Testing
- [ ] End-to-end CLI testing
- [ ] Authentication flow testing  
- [ ] Data fetching validation
- [ ] Error handling verification

#### Documentation
- [ ] Update README with new CLI commands
- [ ] Add usage examples
- [ ] Document configuration options
- [ ] Create troubleshooting guide

#### Performance Testing
- [ ] Concurrent operation testing
- [ ] Memory usage validation
- [ ] Download performance testing
- [ ] Large dataset handling

---

## Phase 1 Deliverables Checklist

### CLI Tool
- [ ] Complete CLI with all major commands
- [ ] Configuration file support
- [ ] Multiple output formats (JSON, table, CSV)
- [ ] Interactive authentication
- [ ] Progress indicators for long operations

### Core Functionality  
- [ ] Activity listing with filtering
- [ ] Activity detail fetching
- [ ] Activity downloading (GPX, TCX, CSV)
- [ ] All existing health data accessible via CLI
- [ ] VO2 max and heart rate zone data

### Code Quality
- [ ] Reorganized package structure
- [ ] Consistent error handling
- [ ] Comprehensive logging
- [ ] Basic test coverage (>60%)
- [ ] Documentation updated

### User Experience
- [ ] Intuitive command structure
- [ ] Helpful error messages
- [ ] Progress feedback
- [ ] Consistent data formatting
- [ ] Working examples and documentation

---

## Success Criteria

1. **CLI Completeness**: All major Garmin data types accessible via CLI
2. **Usability**: New users can get started within 5 minutes
3. **Reliability**: Commands work consistently without errors
4. **Performance**: Downloads and data fetching perform well
5. **Documentation**: Clear examples and troubleshooting available

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| API endpoint changes | High | Create abstraction layer, add endpoint validation |
| Authentication issues | High | Implement robust error handling and retry logic |
| Download format limitations | Medium | Start with GPX, add others incrementally |
| Performance with large datasets | Medium | Implement pagination and caching |
| Package reorganization complexity | Medium | Do incrementally with thorough testing |

## Dependencies

- Cobra CLI framework
- Garmin Connect API stability
- OAuth flow reliability
- File system permissions for downloads
- Network connectivity for API calls

This phase establishes the foundation for all subsequent development while delivering immediate value to users through a comprehensive CLI tool.