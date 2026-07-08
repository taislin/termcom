package save

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/soldier"
)

const CurrentVersion = 3

type SaveData struct {
	Version        int
	GameTime       time.Time
	Funds          int64
	Paused         bool
	TimeSpeed      int
	AlienActivity  int
	SpeciesSeed    int64
	AlienKnowledge map[string]int
	Base           *BaseSave
	UFOs           []*UFOSave
	Missions       []*MissionSave
}

type BaseSave struct {
	Name              string
	Scientists        int
	Engineers         int
	CompletedResearch []string
	UnlockedWeapons   []string
	UnlockedArmor     []string
	Stores            map[string]int
	Soldiers          []*SoldierSave
	Facilities        []*FacilitySave
	ManufactureQueue  []*ManufJobSave
	ActiveResearch    *ResearchSave
}

type SoldierSave struct {
	Name      string
	Rank      int
	HP        int
	MaxHP     int
	TU        int
	MaxTU     int
	Accuracy  int
	Bravery   int
	Reactions int
	Strength  int
	PsiSkill  int
	PsiStr    int
	Weapon    string
	Armor     string
	Kills     int
	Missions  int
	Wounds    int
	WeaponAmmo int
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
	TypeName string
	X, Y     float64
	Active   bool
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
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
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
	if data.Version < 1 {
		return fmt.Errorf("save version %d is too old or invalid", data.Version)
	}
	if data.Version == 1 {
		migrateV1toV2(data)
	}
	if data.Version == 2 {
		migrateV2toV3(data)
	}
	return nil
}

func migrateV1toV2(data *SaveData) {
	if data.Base != nil {
		for _, s := range data.Base.Soldiers {
			if s.WeaponAmmo == 0 {
				s.WeaponAmmo = 50
			}
		}
	}
	data.Version = 2
}

func migrateV2toV3(data *SaveData) {
	if data.AlienKnowledge == nil {
		data.AlienKnowledge = make(map[string]int)
	}
	data.Version = 3
}

func FromBase(b *base.Base) *BaseSave {
	bs := &BaseSave{
		Name:              b.Name,
		Scientists:        b.Scientists,
		Engineers:         b.Engineers,
		CompletedResearch: b.CompletedResearch,
		UnlockedWeapons:   b.UnlockedWeapons,
		UnlockedArmor:     b.UnlockedArmor,
		Stores:            b.Stores,
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
			WeaponAmmo: s.WeaponAmmo,
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
	b := base.NewBase(bs.Name)
	b.Soldiers = nil
	b.Scientists = bs.Scientists
	b.Engineers = bs.Engineers
	b.CompletedResearch = bs.CompletedResearch
	b.UnlockedWeapons = bs.UnlockedWeapons
	b.UnlockedArmor = bs.UnlockedArmor
	b.Stores = bs.Stores
	if b.Stores == nil {
		b.Stores = make(map[string]int)
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
