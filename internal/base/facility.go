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
	Stores      map[string]int
}

func NewBase(name string) *Base {
	return &Base{
		Name:       name,
		Scientists: 10,
		Engineers:  10,
		MaxStorage: 50,
		Stores:     make(map[string]int),
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

func (b *Base) AddItem(item string, qty int) {
	b.Stores[item] += qty
}

func (b *Base) RemoveItem(item string, qty int) bool {
	if b.Stores[item] < qty {
		return false
	}
	b.Stores[item] -= qty
	if b.Stores[item] == 0 {
		delete(b.Stores, item)
	}
	return true
}

func (b *Base) CountItem(item string) int {
	return b.Stores[item]
}

func (b *Base) AddLoot(items []string) {
	for _, item := range items {
		b.AddItem(item, 1)
	}
}

func (b *Base) EquipWeapon(soldierIdx int, weaponKey string) bool {
	if soldierIdx < 0 || soldierIdx >= len(b.Soldiers) {
		return false
	}
	if b.CountItem(weaponKey) <= 0 {
		return false
	}
	s := b.Soldiers[soldierIdx]
	if s.Weapon != "" && s.Weapon != "pistol" {
		b.AddItem(s.Weapon, 1)
	}
	if !b.RemoveItem(weaponKey, 1) {
		return false
	}
	s.Weapon = weaponKey
	return true
}

func (b *Base) EquipArmor(soldierIdx int, armorKey string) bool {
	if soldierIdx < 0 || soldierIdx >= len(b.Soldiers) {
		return false
	}
	if armorKey != "none" && b.CountItem(armorKey) <= 0 {
		return false
	}
	s := b.Soldiers[soldierIdx]
	if s.Armor != "none" {
		b.AddItem(s.Armor, 1)
	}
	if armorKey != "none" {
		if !b.RemoveItem(armorKey, 1) {
			return false
		}
	}
	s.Armor = armorKey
	return true
}

func (b *Base) MonthlySalary() int {
	return (len(b.Soldiers) + b.Scientists + b.Engineers) * 2000
}

func (b *Base) GovernmentFunding() int {
	radarCount := b.CountFacility(FacRadar)
	baseFunding := 200000
	return baseFunding + radarCount*50000
}

func (b *Base) AdvanceMonth() (salary, funding int) {
	salary = b.MonthlySalary()
	funding = b.GovernmentFunding()
	return salary, funding
}

func (b *Base) StorageCapacity() int {
	return b.CountFacility(FacStorage) * 50
}

func (b *Base) TotalWeight() int {
	total := 0
	for item, qty := range b.Stores {
		if it, ok := data.Items[item]; ok {
			total += it.Weight * qty
		} else if w, ok := data.Weapons[item]; ok {
			total += 5 * qty
			_ = w
		} else if _, ok := data.Armors[item]; ok {
			total += 8 * qty
		}
	}
	return total
}
