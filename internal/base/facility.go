package base

import (
	"fmt"
	"math/rand"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
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
	BuildDays int
}

var FacilityDefs = map[FacilityType]FacilityInfo{
	FacLivingQuarters: {Name: "Living Quarters", Short: "LQ", Cost: 50000, BuildDays: 5},
	FacLab:            {Name: "Laboratory", Short: "LAB", Cost: 75000, BuildDays: 7},
	FacWorkshop:       {Name: "Workshop", Short: "WRK", Cost: 60000, BuildDays: 7},
	FacStorage:        {Name: "Storage", Short: "STR", Cost: 40000, BuildDays: 3},
	FacRadar:          {Name: "Radar", Short: "RAD", Cost: 80000, BuildDays: 5},
	FacContainment:    {Name: "Alien Containment", Short: "ALC", Cost: 100000, BuildDays: 10},
	FacPsiLab:         {Name: "Psi-Lab", Short: "PSI", Cost: 150000, BuildDays: 14},
	FacHangar:         {Name: "Hangar", Short: "HNG", Cost: 120000, BuildDays: 8},
}

var facilityNameKeys = map[FacilityType]string{
	FacLivingQuarters: "FAC_LIVING_QUARTERS",
	FacLab:            "FAC_LABORATORY",
	FacWorkshop:       "FAC_WORKSHOP",
	FacStorage:        "FAC_STORAGE",
	FacRadar:          "FAC_RADAR",
	FacContainment:    "FAC_ALIEN_CONTAINMENT",
	FacPsiLab:         "FAC_PSI_LAB",
	FacHangar:         "FAC_HANGAR",
}

func FacilityDisplayName(ft FacilityType) string {
	if key, ok := facilityNameKeys[ft]; ok {
		return language.String(key)
	}
	return FacilityDefs[ft].Name
}

type Facility struct {
	Type     FacilityType
	Building bool
	DaysLeft int
	Row      int
	Col      int
}

const HireCost = 50000

type ResearchProject struct {
	TopicID    string
	Progress   int
	Cost       int
	Scientists int
	Completed  bool
}

type ManufactureJob struct {
	ItemKey   string
	Count     int
	Progress  int
	CostDays  int
	Materials map[string]int
	Engineers int
	Completed bool
}

func (j *ManufactureJob) DisplayName() string {
	if r, ok := data.RuleItems[j.ItemKey]; ok {
		return r.DisplayName()
	}
	if a, ok := data.Armors[j.ItemKey]; ok {
		return a.DisplayNameByKey(j.ItemKey)
	}
	return j.ItemKey
}

type Base struct {
	Name                 string
	CityID               int // geoscape city ID where this base is located
	Facilities           []*Facility
	Soldiers             []*soldier.Soldier
	Scientists           int
	Engineers            int
	UnassignedScientists int
	UnassignedEngineers  int
	UsedStorage          int
	Stores               map[string]int
	CompletedResearch    []string
	ActiveResearch       *ResearchProject
	ManufactureQueue     []*ManufactureJob
	UnlockedWeapons      []string
	UnlockedArmor        []string
	Hangars              []*data.InterceptorState // Manage interceptors here
	CustomWeapons        map[string]*data.WeaponDesign // player-designed weapons
	LiveAliens           []string                 // Captured aliens
	AlienActivity        int
}

func NewBase(name string, cityID int) *Base {
	b := &Base{
		Name:                 name,
		CityID:               cityID,
		Scientists:           10,
		UnassignedScientists: 10,
		Engineers:            10,
		UnassignedEngineers:  10,
		Stores:               make(map[string]int),
		Hangars:              make([]*data.InterceptorState, 0),
		CustomWeapons:        make(map[string]*data.WeaponDesign),
		LiveAliens:           make([]string, 0),
	}
	for i := 0; i < 4; i++ {
		s := soldier.NewSoldier(soldier.RandomName())
		b.Soldiers = append(b.Soldiers, s)
	}
	// Start with 1 hangar and 1 interceptor
	b.Facilities = append(b.Facilities, &Facility{Type: FacHangar, Row: 0, Col: 0})
	w := data.InterceptorWeapons["avalanche"]
	b.Hangars = append(b.Hangars, &data.InterceptorState{
		ID:        0,
		Name:      language.String("INTERCEPTOR_DEFAULT_NAME"),
		WeaponKey: "avalanche",
		HP:        60,
		MaxHP:     60,
		Ammo:      w.FireRate * 4,
		Status:    "available",
	})
	return b
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

func (b *Base) BuyInterceptor(weaponKey string, funds *int64) bool {
	w, ok := data.InterceptorWeapons[weaponKey]
	if !ok {
		return false
	}
	cost := int64(w.Cost)
	if *funds < cost {
		return false
	}
	// Replace a destroyed interceptor first
	for i, h := range b.Hangars {
		if h.Status == "destroyed" {
			b.Hangars[i] = &data.InterceptorState{
				ID:        h.ID,
				Name:      language.String("INTERCEPTOR_DEFAULT_NAME"),
				WeaponKey: weaponKey,
				HP:        60,
				MaxHP:     60,
				Ammo:      w.FireRate * 4,
				Status:    "available",
			}
			*funds -= cost
			return true
		}
	}
	hangarCount := b.CountFacility(FacHangar)
	if len(b.Hangars) >= hangarCount {
		return false
	}
	*funds -= cost
	b.Hangars = append(b.Hangars, &data.InterceptorState{
		ID:        len(b.Hangars),
		Name:      language.String("INTERCEPTOR_DEFAULT_NAME"),
		WeaponKey: weaponKey,
		HP:        60,
		MaxHP:     60,
		Ammo:      w.FireRate * 4,
		Status:    "available",
	})
	return true
}

func (b *Base) GetAvailableInterceptors() []*data.InterceptorState {
	var available []*data.InterceptorState
	for _, h := range b.Hangars {
		if h.Status == "available" {
			available = append(available, h)
		}
	}
	return available
}

var interceptorWeaponOrder = []string{"avalanche", "stingray", "cannon"}

func (b *Base) ChangeInterceptorWeapon(idx int) string {
	if idx < 0 || idx >= len(b.Hangars) {
		return ""
	}
	hg := b.Hangars[idx]
	if hg.Status != "available" {
		return ""
	}
	curIdx := -1
	for i, k := range interceptorWeaponOrder {
		if k == hg.WeaponKey {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + 1) % len(interceptorWeaponOrder)
	newKey := interceptorWeaponOrder[nextIdx]
	w := data.InterceptorWeapons[newKey]
	hg.WeaponKey = newKey
	hg.Ammo = w.FireRate * 4
	return w.Name
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

// HealthySoldiers returns soldiers fit for deployment (HP > 0, no wounds, no fatigue).
func (b *Base) HealthySoldiers() []*soldier.Soldier {
	var healthy []*soldier.Soldier
	for _, s := range b.Soldiers {
		if s.CanDeploy() {
			healthy = append(healthy, s)
		}
	}
	return healthy
}

func (b *Base) HireSoldier() (bool, string) {
	cap := b.LivingCapacity()
	if len(b.Soldiers) >= cap {
		return false, language.String("MSG_NO_ROOM")
	}
	s := soldier.NewSoldier(soldier.RandomName())
	b.Soldiers = append(b.Soldiers, s)
	return true, fmt.Sprintf(language.String("MSG_HIRED"), s.Name)
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

// AdjacentCount returns how many completed (non-building) facilities of type
// targetType are orthogonally adjacent to any facility of type ft in this base.
// Checks the Row/Col grid (8-column layout).
func (b *Base) AdjacentCount(ft, targetType FacilityType) int {
	// Build a set of occupied grid positions for the target facility type
	type pos struct{ row, col int }
	targets := make(map[pos]bool)
	for _, f := range b.Facilities {
		if !f.Building && f.Type == targetType {
			targets[pos{f.Row, f.Col}] = true
		}
	}
	count := 0
	for _, f := range b.Facilities {
		if f.Building || f.Type != ft {
			continue
		}
		for _, d := range []pos{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			if targets[pos{f.Row + d.row, f.Col + d.col}] {
				count++
			}
		}
	}
	return count
}

// AdjacentResearchBonus returns a research speed multiplier (1.0 = no bonus)
// based on Lab-to-Lab adjacency. Each adjacent Lab gives +10% (max +30%).
func (b *Base) AdjacentResearchBonus() float64 {
	n := b.AdjacentCount(FacLab, FacLab)
	if n > 3 {
		n = 3
	}
	return 1.0 + float64(n)*0.10
}

// AdjacentManufactureBonus returns a manufacture speed multiplier (1.0 = no bonus)
// based on Workshop-to-Workshop adjacency. Each adjacent Workshop gives +10% (max +30%).
func (b *Base) AdjacentManufactureBonus() float64 {
	n := b.AdjacentCount(FacWorkshop, FacWorkshop)
	if n > 3 {
		n = 3
	}
	return 1.0 + float64(n)*0.10
}

// AdjacentHealBonus returns extra HP healed per day from adjacent Living Quarters.
// Each adjacent pair provides +1 HP/day (max +3).
func (b *Base) AdjacentHealBonus() int {
	n := b.AdjacentCount(FacLivingQuarters, FacLivingQuarters)
	if n > 3 {
		n = 3
	}
	return n
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
	hasPsiLab := b.CountFacility(FacPsiLab) > 0
	for _, s := range b.Soldiers {
		if s.Wounds > 0 {
			s.Wounds--
			if s.Wounds <= 0 {
				s.Wounds = 0
				s.HP = s.MaxHP
			} else {
				healRate := 2 + b.AdjacentHealBonus()
				s.HP += healRate
				if s.HP > s.MaxHP {
					s.HP = s.MaxHP
				}
			}
		}
		if s.Fatigue > 0 {
			s.Fatigue--
		}
		if hasPsiLab && s.Wounds <= 0 && s.PsiSkill > 0 && s.PsiSkill < 80 {
			if rand.Intn(100) < 8 {
				s.PsiSkill++
			}
		}
	}
}

func GetBuildableWeapons() []string {
	var items []string
	for k, w := range data.RuleItems {
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

func (b *Base) AddItem(item string, qty int) bool {
	if b.UsedStorage+qty > b.StorageCapacity() {
		return false
	}
	b.Stores[item] += qty
	b.UsedStorage += qty
	return true
}

func (b *Base) RemoveItem(item string, qty int) bool {
	if b.Stores[item] < qty {
		return false
	}
	b.Stores[item] -= qty
	b.UsedStorage -= qty
	if b.Stores[item] == 0 {
		delete(b.Stores, item)
	}
	return true
}

func (b *Base) CountItem(item string) int {
	return b.Stores[item]
}

func (b *Base) SellItem(item string) int64 {
	if b.Stores[item] <= 0 {
		return 0
	}
	var value int64
	if ri, ok := data.RuleItems[item]; ok && ri.CostSell > 0 {
		value = int64(ri.CostSell)
	} else if it, ok := data.Items[item]; ok && it.Value > 0 {
		value = int64(it.Value)
	} else {
		value = 2000
	}
	b.RemoveItem(item, 1)
	return value
}

// AddLoot stores items in the base, returning the number that could not be
// stored because storage was full.
func (b *Base) AddLoot(items []string) int {
	dropped := 0
	for _, item := range items {
		if !b.AddItem(item, 1) {
			dropped++
		}
	}
	return dropped
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
		if !b.AddItem(s.Weapon, 1) {
			return false
		}
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
	if armorKey != "none" {
		if !b.RemoveItem(armorKey, 1) {
			return false
		}
	}
	if s.Armor != "none" {
		if !b.AddItem(s.Armor, 1) {
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
	baseFunding := 300000
	return baseFunding + radarCount*75000
}

func (b *Base) AdvanceMonth() (salary, funding int) {
	salary = b.MonthlySalary()
	funding = b.GovernmentFunding()

	// Adjust funding based on AlienActivity
	// High activity (near 100) reduces funding significantly.
	// Activity 0: 100% funding
	// Activity 100: ~40% funding
	activityPenalty := float64(b.AlienActivity) * 0.6 / 100.0
	funding = int(float64(funding) * (1.0 - activityPenalty))

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
		} else if w, ok := data.RuleItems[item]; ok {
			total += w.Weight * qty
		} else if _, ok := data.Armors[item]; ok {
			total += 8 * qty
		}
	}
	return total
}

// InterrogateAlien consumes a captured alien from LiveAliens and grants a
// research bonus: auto-completes the matching autopsy topic or adds large
// progress to the active research if it matches. Returns the topic name
// completed/bonused and whether the interrogation succeeded.
func (b *Base) InterrogateAlien(alienName string) (string, bool) {
	if len(b.LiveAliens) == 0 || b.TotalLabs() == 0 {
		return "", false
	}
	// Find alien in LiveAliens
	idx := -1
	for i, a := range b.LiveAliens {
		if a == alienName {
			idx = i
			break
		}
	}
	if idx < 0 {
		return "", false
	}

	// Validate we can provide a benefit BEFORE consuming the alien
	at := data.GetAlienByName(alienName)
	if at == nil {
		return "", false
	}
	autopsyID := at.AutopsyID
	if autopsyID == "" {
		return "", false
	}
	topic := data.ResearchByID(autopsyID)
	if topic == nil {
		return "", false
	}

	// Determine benefit before consuming
	benefit := false
	topicName := ""
	if b.HasResearch(autopsyID) {
		if b.ActiveResearch != nil && !b.ActiveResearch.Completed {
			b.ActiveResearch.Progress += b.ActiveResearch.Cost / 4
			benefit = true
			topicName = topic.DisplayName()
		}
	} else if b.ActiveResearch != nil && b.ActiveResearch.TopicID == autopsyID && !b.ActiveResearch.Completed {
		b.ActiveResearch.Progress = b.ActiveResearch.Cost
		benefit = true
		topicName = topic.DisplayName()
	} else {
		b.CompletedResearch = append(b.CompletedResearch, autopsyID)
		b.UnlockedWeapons = append(b.UnlockedWeapons, topic.UnlockWeap...)
		b.UnlockedArmor = append(b.UnlockedArmor, topic.UnlockArmor...)
		for _, item := range topic.UnlockItems {
			b.AddItem(item, 1)
		}
		for _, wpn := range topic.UnlockWeap {
			if _, ok := data.RuleItems[wpn]; ok {
				b.AddItem(wpn, 1)
			}
		}
		for _, arm := range topic.UnlockArmor {
			if _, ok := data.Armors[arm]; ok {
				b.AddItem(arm, 1)
			}
		}
		benefit = true
		topicName = topic.DisplayName()
	}

	if !benefit {
		return "", false
	}

	// Now consume the alien
	b.LiveAliens = append(b.LiveAliens[:idx], b.LiveAliens[idx+1:]...)
	return topicName, true
}

func (b *Base) HasResearch(topicID string) bool {
	for _, id := range b.CompletedResearch {
		if id == topicID {
			return true
		}
	}
	return false
}

func (b *Base) CanResearch(topic *data.ResearchTopic) bool {
	if b.HasResearch(topic.ID) {
		return false
	}
	if b.TotalLabs() == 0 {
		return false
	}
	for _, req := range topic.Requires {
		if !b.HasResearch(req) {
			return false
		}
	}
	return true
}

func (b *Base) StartResearch(topicID string) bool {
	topic := data.ResearchByID(topicID)
	if topic == nil || !b.CanResearch(topic) {
		return false
	}
	if b.ActiveResearch != nil && !b.ActiveResearch.Completed {
		return false
	}
	b.ActiveResearch = &ResearchProject{
		TopicID:    topicID,
		Cost:       topic.Cost,
		Scientists: 0,
	}
	return true
}

func (b *Base) AssignScientists(scientists int) bool {
	if b.ActiveResearch == nil || b.ActiveResearch.Completed {
		return false
	}
	if scientists < 0 {
		// Unassign
		amount := -scientists
		if amount > b.ActiveResearch.Scientists {
			amount = b.ActiveResearch.Scientists
		}
		b.ActiveResearch.Scientists -= amount
		b.UnassignedScientists += amount
	} else {
		// Assign
		amount := scientists
		if amount > b.UnassignedScientists {
			amount = b.UnassignedScientists
		}
		b.ActiveResearch.Scientists += amount
		b.UnassignedScientists -= amount
	}
	return true
}

func (b *Base) AdvanceResearch() []string {
	if b.ActiveResearch == nil || b.ActiveResearch.Completed || b.ActiveResearch.Scientists == 0 {
		return nil
	}
	bonus := b.AdjacentResearchBonus()
	b.ActiveResearch.Progress += int(float64(b.ActiveResearch.Scientists) * bonus)
	if b.ActiveResearch.Progress >= b.ActiveResearch.Cost {
		b.ActiveResearch.Completed = true
		topic := data.ResearchByID(b.ActiveResearch.TopicID)
		b.CompletedResearch = append(b.CompletedResearch, b.ActiveResearch.TopicID)
		name := b.ActiveResearch.TopicID
		if topic != nil {
			b.UnlockedWeapons = append(b.UnlockedWeapons, topic.UnlockWeap...)
			b.UnlockedArmor = append(b.UnlockedArmor, topic.UnlockArmor...)
			for _, item := range topic.UnlockItems {
				b.AddItem(item, 1)
			}
			for _, wpn := range topic.UnlockWeap {
				if _, ok := data.RuleItems[wpn]; ok {
					b.AddItem(wpn, 1)
				}
			}
			for _, arm := range topic.UnlockArmor {
				if _, ok := data.Armors[arm]; ok {
					b.AddItem(arm, 1)
				}
			}
			if b.ActiveResearch.TopicID == "mind_control" {
				for _, s := range b.Soldiers {
					s.PsiSkill += 20
					if s.PsiSkill > 100 {
						s.PsiSkill = 100
					}
				}
			}
			name = topic.DisplayName()
		}
		b.ActiveResearch = nil
		return []string{name}
	}
	return nil
}

func (b *Base) CanManufacture(item string, count int) bool {
	if count <= 0 {
		return false
	}
	if b.TotalWorkshops() == 0 {
		return false
	}
	if b.Engineers <= 0 {
		return false
	}
	return true
}

func (b *Base) StartManufacture(item string, count int, materials map[string]int) bool {
	if !b.CanManufacture(item, count) {
		return false
	}
	for mat, qty := range materials {
		if b.CountItem(mat) < qty*count {
			return false
		}
	}
	for mat, qty := range materials {
		b.RemoveItem(mat, qty*count)
	}
	job := &ManufactureJob{
		ItemKey:   item,
		Count:     count,
		CostDays:  5 + count*2,
		Materials: materials,
		Engineers: 0,
	}
	b.ManufactureQueue = append(b.ManufactureQueue, job)
	return true
}

func (b *Base) AssignEngineers(jobIdx, engineers int) bool {
	if jobIdx < 0 || jobIdx >= len(b.ManufactureQueue) {
		return false
	}
	job := b.ManufactureQueue[jobIdx]
	if engineers < 0 {
		// Unassign
		amount := -engineers
		if amount > job.Engineers {
			amount = job.Engineers
		}
		job.Engineers -= amount
		b.UnassignedEngineers += amount
	} else {
		// Assign
		amount := engineers
		if amount > b.UnassignedEngineers {
			amount = b.UnassignedEngineers
		}
		job.Engineers += amount
		b.UnassignedEngineers -= amount
	}
	return true
}

func (b *Base) AdvanceManufacture() []string {
	var completed []string
	bonus := b.AdjacentManufactureBonus()
	remaining := make([]*ManufactureJob, 0, len(b.ManufactureQueue))
	for _, job := range b.ManufactureQueue {
		if job.Completed {
			continue
		}
		if job.Engineers > 0 {
			job.Progress += int(float64(job.Engineers) * bonus)
		}
		if job.Progress >= job.CostDays {
			b.AddItem(job.ItemKey, job.Count)
			// Add to unlocked list if not already there
			found := false
			for _, w := range b.UnlockedWeapons {
				if w == job.ItemKey {
					found = true
					break
				}
			}
			if !found {
				b.UnlockedWeapons = append(b.UnlockedWeapons, job.ItemKey)
			}
			completed = append(completed, job.ItemKey)
			job.Completed = true
		} else {
			remaining = append(remaining, job)
		}
	}
	b.ManufactureQueue = remaining
	return completed
}
