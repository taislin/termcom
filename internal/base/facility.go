package base

import (
	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/soldier"
)

type FacilityType int

const (
	FacLivingQuarters FacilityType = iota
	FacLab
	FacWorkshop
	FacStorage
	FacRadar
	FacContainment
	FacPsiLab
	FacHangar
)

type FacilityInfo struct {
	Name      string
	Short     string
	Cost      int
	Size      int
	BuildDays int
}

var FacilityDefs = map[FacilityType]FacilityInfo{
	FacLivingQuarters: {Name: "Living Quarters", Short: "LQ", Cost: 50000, Size: 1, BuildDays: 5},
	FacLab:            {Name: "Laboratory", Short: "LAB", Cost: 75000, Size: 1, BuildDays: 7},
	FacWorkshop:       {Name: "Workshop", Short: "WRK", Cost: 60000, Size: 1, BuildDays: 7},
	FacStorage:        {Name: "Storage", Short: "STR", Cost: 40000, Size: 1, BuildDays: 3},
	FacRadar:          {Name: "Radar", Short: "RAD", Cost: 80000, Size: 1, BuildDays: 5},
	FacContainment:    {Name: "Alien Containment", Short: "ALC", Cost: 100000, Size: 1, BuildDays: 10},
	FacPsiLab:         {Name: "Psi-Lab", Short: "PSI", Cost: 150000, Size: 1, BuildDays: 14},
	FacHangar:         {Name: "Hangar", Short: "HNG", Cost: 120000, Size: 1, BuildDays: 8},
}

type Facility struct {
	Type     FacilityType
	Building bool
	DaysLeft int
	Row      int
	Col      int
}

const HireCost = 50000

type Base struct {
	Name        string
	Facilities  []*Facility
	Soldiers    []*soldier.Soldier
	Scientists  int
	Engineers   int
	MaxStorage  int
	UsedStorage int
}

func NewBase(name string) *Base {
	return &Base{
		Name:       name,
		Scientists: 10,
		Engineers:  10,
		MaxStorage: 50,
	}
}

func (b *Base) CountFacility(ft FacilityType) int {
	count := 0
	for _, f := range b.Facilities {
		if f.Type == ft && !f.Building {
			count++
		}
	}
	return count
}

func (b *Base) LivingCapacity() int {
	return b.CountFacility(FacLivingQuarters) * 8
}

func (b *Base) BuildFacility(ft FacilityType) bool {
	def := FacilityDefs[ft]
	fac := &Facility{
		Type:     ft,
		Building: true,
		DaysLeft: def.BuildDays,
		Row:      len(b.Facilities) / 8,
		Col:      len(b.Facilities) % 8,
	}
	b.Facilities = append(b.Facilities, fac)
	return true
}

func (b *Base) HireSoldier() (bool, string) {
	cap := b.LivingCapacity()
	if len(b.Soldiers) >= cap {
		return false, "No room! Build more Living Quarters."
	}
	s := soldier.NewSoldier(soldier.RandomName())
	b.Soldiers = append(b.Soldiers, s)
	return true, s.Name + " hired."
}

func (b *Base) DismissSoldier(idx int) bool {
	if idx < 0 || idx >= len(b.Soldiers) {
		return false
	}
	b.Soldiers = append(b.Soldiers[:idx], b.Soldiers[idx+1:]...)
	return true
}

func (b *Base) RemoveDeadSoldiers() []string {
	var names []string
	alive := make([]*soldier.Soldier, 0, len(b.Soldiers))
	for _, s := range b.Soldiers {
		if s.HP <= 0 {
			names = append(names, s.Name)
		} else {
			alive = append(alive, s)
		}
	}
	b.Soldiers = alive
	return names
}

func (b *Base) TotalLabs() int {
	return b.CountFacility(FacLab)
}

func (b *Base) TotalWorkshops() int {
	return b.CountFacility(FacWorkshop)
}

func (b *Base) AdvanceDay() {
	for _, f := range b.Facilities {
		if f.Building {
			f.DaysLeft--
			if f.DaysLeft <= 0 {
				f.Building = false
			}
		}
	}
}

type ManufactureItem struct {
	Item     string
	Count    int
	Assigned int
	DaysLeft int
	Queue    int
}

func GetBuildableWeapons() []string {
	var items []string
	for k, w := range data.Weapons {
		if !w.IsAlien {
			items = append(items, k)
		}
	}
	return items
}

func GetBuildableArmor() []string {
	var items []string
	for k := range data.Armors {
		if k != "none" {
			items = append(items, k)
		}
	}
	return items
}
