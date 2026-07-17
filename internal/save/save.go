package save

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
)

const (
	CurrentVersion = 4
	saveFilePerm   = 0644 // owner read/write, group/other read
	maxSlots       = 10
	fundsDivK      = 1000
)

type AlienBaseSave struct {
	CityID          int
	Threat          int
	TurnsAlive      int
	LastMissionTick int
	DefendingUFOID  int
	Name            string
}

type SaveData struct {
	Version        int
	Slot           int
	GameTime       time.Time
	Funds          int64
	Paused         bool
	TimeSpeed      int
	Difficulty     int
	AlienActivity  int
	SpeciesSeed    int64
	AlienKnowledge map[string]int
	Bases          []*BaseSave
	UFOs           []*UFOSave
	Missions       []*MissionSave
	MissionsWon    int
	AlienBases     []*AlienBaseSave
}

type BaseSave struct {
	Name                 string
	CityID               int
	Scientists           int
	Engineers            int
	UnassignedScientists int
	UnassignedEngineers  int
	CompletedResearch    []string
	UnlockedWeapons      []string
	UnlockedArmor        []string
	Stores               map[string]int
	UsedStorage          int
	LiveAliens           []string
	Soldiers             []*SoldierSave
	Facilities           []*FacilitySave
	ManufactureQueue     []*ManufJobSave
	ActiveResearch       *ResearchSave
	Hangars              []*data.InterceptorState
}

type SoldierSave struct {
	Name       string
	Rank       int
	HP         int
	MaxHP      int
	TU         int
	MaxTU      int
	Accuracy   int
	Bravery    int
	Reactions  int
	Strength   int
	PsiSkill   int
	PsiStr     int
	Weapon     string
	Armor      string
	Kills      int
	Missions   int
	Wounds     int
	Fatigue    int
	WeaponAmmo int
	Perks      []string
}

type FacilitySave struct {
	Type     int
	Building bool
	DaysLeft int
	Row      int
	Col      int
}

type ManufJobSave struct {
	ItemKey   string
	Count     int
	Progress  int
	CostDays  int
	Materials map[string]int
	Engineers int
	Completed bool
}

type ResearchSave struct {
	TopicID    string
	Progress   int
	Cost       int
	Scientists int
	Completed  bool
}

type UFOSave struct {
	ID        int
	TypeName  string
	X, Y      float64
	Progress  float64
	NodeFrom  int
	NodeTo    int
	TurnsLeft int
	Active    bool
}

type MissionSave struct {
	Type      string
	CityName  string
	TurnsLeft int
	HoursLeft float64
	X, Y      int
}

func SaveGame(path string, data *SaveData) error {
	data.Version = CurrentVersion
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// Write to a temp file first, then rename, so a failure never leaves a
	// truncated/corrupt save behind.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf, saveFilePerm); err != nil {
		return err
	}
	os.Remove(path)
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func SaveGameToSlot(slot int, data *SaveData) error {
	data.Slot = slot
	return SaveGame(SavePath(slot), data)
}

func LoadGame(path string) (*SaveData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var data SaveData
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}
	if err := migrateSave(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func migrateSave(data *SaveData) error {
	if data.Version > CurrentVersion {
		return fmt.Errorf("save version %d is newer than current version %d", data.Version, CurrentVersion)
	}
	if data.Version < 2 {
		return fmt.Errorf("save version %d is too old (minimum v2)", data.Version)
	}
	if data.Version == 2 {
		migrateV2toV3(data)
	}
	if data.Version == 3 {
		migrateV3toV4(data)
	}
	return nil
}

func migrateV3toV4(data *SaveData) {
	if data.AlienBases == nil {
		data.AlienBases = make([]*AlienBaseSave, 0)
	}
	data.Version = 4
}

func migrateV2toV3(data *SaveData) {
	if data.AlienKnowledge == nil {
		data.AlienKnowledge = make(map[string]int)
	}
	data.Version = 3
}

func SavePath(slot int) string {
	if slot == 0 {
		return "xcom_save.json"
	}
	return fmt.Sprintf("save_slot_%d.json", slot)
}

func AutoSavePath() string {
	return "autosave.json"
}

func ListSlots() []int {
	var slots []int
	for slot := 1; slot <= maxSlots; slot++ {
		if _, err := os.Stat(SavePath(slot)); err == nil {
			slots = append(slots, slot)
		}
	}
	return slots
}

func LoadSaveInfo(slot int) (string, error) {
	sd, err := LoadGame(SavePath(slot))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(language.String("SLOT_FORMAT"), slot, sd.GameTime.Format("2006 Jan 02"), sd.Funds/fundsDivK), nil
}

func FromBase(b *base.Base) *BaseSave {
	bs := &BaseSave{
		Name:                 b.Name,
		CityID:               b.CityID,
		Scientists:           b.Scientists,
		Engineers:            b.Engineers,
		UnassignedScientists: b.UnassignedScientists,
		UnassignedEngineers:  b.UnassignedEngineers,
		CompletedResearch:    b.CompletedResearch,
		UnlockedWeapons:      b.UnlockedWeapons,
		UnlockedArmor:        b.UnlockedArmor,
		Stores:               b.Stores,
		UsedStorage:          b.UsedStorage,
		LiveAliens:           b.LiveAliens,
		Hangars:              b.Hangars,
	}
	for _, s := range b.Soldiers {
		bs.Soldiers = append(bs.Soldiers, &SoldierSave{
			Name:       s.Name,
			Rank:       int(s.Rank),
			HP:         s.HP,
			MaxHP:      s.MaxHP,
			TU:         s.TU,
			MaxTU:      s.MaxTU,
			Accuracy:   s.Accuracy,
			Bravery:    s.Bravery,
			Reactions:  s.Reactions,
			Strength:   s.Strength,
			PsiSkill:   s.PsiSkill,
			PsiStr:     s.PsiStr,
			Weapon:     s.Weapon,
			Armor:      s.Armor,
			Kills:      s.Kills,
			Missions:   s.Missions,
			Wounds:     s.Wounds,
			Fatigue:    s.Fatigue,
			WeaponAmmo: s.WeaponAmmo,
			Perks:      s.Perks,
		})
	}
	for _, f := range b.Facilities {
		bs.Facilities = append(bs.Facilities, &FacilitySave{
			Type:     int(f.Type),
			Building: f.Building,
			DaysLeft: f.DaysLeft,
			Row:      f.Row,
			Col:      f.Col,
		})
	}
	for _, j := range b.ManufactureQueue {
		bs.ManufactureQueue = append(bs.ManufactureQueue, &ManufJobSave{
			ItemKey:   j.ItemKey,
			Count:     j.Count,
			Progress:  j.Progress,
			CostDays:  j.CostDays,
			Materials: j.Materials,
			Engineers: j.Engineers,
			Completed: j.Completed,
		})
	}
	if b.ActiveResearch != nil {
		bs.ActiveResearch = &ResearchSave{
			TopicID:    b.ActiveResearch.TopicID,
			Progress:   b.ActiveResearch.Progress,
			Cost:       b.ActiveResearch.Cost,
			Scientists: b.ActiveResearch.Scientists,
			Completed:  b.ActiveResearch.Completed,
		}
	}
	return bs
}

func ToBase(bs *BaseSave) *base.Base {
	b := base.NewBase(bs.Name, bs.CityID)
	b.Soldiers = nil
	b.Facilities = nil
	b.Hangars = nil
	b.Scientists = bs.Scientists
	b.Engineers = bs.Engineers
	b.UnassignedScientists = bs.UnassignedScientists
	b.UnassignedEngineers = bs.UnassignedEngineers
	b.CompletedResearch = bs.CompletedResearch
	b.UnlockedWeapons = bs.UnlockedWeapons
	b.UnlockedArmor = bs.UnlockedArmor
	b.Stores = bs.Stores
	if b.Stores == nil {
		b.Stores = make(map[string]int)
	}
	b.UsedStorage = bs.UsedStorage
	b.LiveAliens = bs.LiveAliens
	if b.LiveAliens == nil {
		b.LiveAliens = make([]string, 0)
	}
	b.Hangars = bs.Hangars
	if b.Hangars == nil {
		b.Hangars = make([]*data.InterceptorState, 0)
	}
	for _, ss := range bs.Soldiers {
		s := soldier.NewSoldier(ss.Name)
		s.Rank = soldier.Rank(ss.Rank)
		s.HP = ss.HP
		s.MaxHP = ss.MaxHP
		s.TU = ss.TU
		s.MaxTU = ss.MaxTU
		s.Accuracy = ss.Accuracy
		s.Bravery = ss.Bravery
		s.Reactions = ss.Reactions
		s.Strength = ss.Strength
		s.PsiSkill = ss.PsiSkill
		s.PsiStr = ss.PsiStr
		s.Weapon = ss.Weapon
		s.Armor = ss.Armor
		s.Kills = ss.Kills
		s.Missions = ss.Missions
		s.Wounds = ss.Wounds
		s.Fatigue = ss.Fatigue
		s.Perks = ss.Perks
		s.WeaponAmmo = ss.WeaponAmmo
		b.Soldiers = append(b.Soldiers, s)
	}
	for _, fs := range bs.Facilities {
		b.Facilities = append(b.Facilities, &base.Facility{
			Type:     base.FacilityType(fs.Type),
			Building: fs.Building,
			DaysLeft: fs.DaysLeft,
			Row:      fs.Row,
			Col:      fs.Col,
		})
	}
	for _, js := range bs.ManufactureQueue {
		b.ManufactureQueue = append(b.ManufactureQueue, &base.ManufactureJob{
			ItemKey:   js.ItemKey,
			Count:     js.Count,
			Progress:  js.Progress,
			CostDays:  js.CostDays,
			Materials: js.Materials,
			Engineers: js.Engineers,
			Completed: js.Completed,
		})
	}
	if bs.ActiveResearch != nil {
		b.ActiveResearch = &base.ResearchProject{
			TopicID:    bs.ActiveResearch.TopicID,
			Progress:   bs.ActiveResearch.Progress,
			Cost:       bs.ActiveResearch.Cost,
			Scientists: bs.ActiveResearch.Scientists,
			Completed:  bs.ActiveResearch.Completed,
		}
	}
	return b
}
