package data

type ResearchTopic struct {
	ID          string
	Name        string
	Cost        int // man-days
	Requires    []string
	UnlockItems []string
	UnlockWeap  []string
	UnlockArmor []string
	AlienLore   bool
}

var ResearchTree = []ResearchTopic{
	// Basic analysis
	{ID: "alien_alloys", Name: "Alien Alloys", Cost: 60, UnlockItems: []string{"alloys"}},
	{ID: "elerium", Name: "Elerium-115", Cost: 80, UnlockItems: []string{"elerium"}},
	{ID: "sectoid_autopsy", Name: "Sectoid Autopsy", Cost: 40, AlienLore: true},
	{ID: "floater_autopsy", Name: "Floater Autopsy", Cost: 50, AlienLore: true},
	{ID: "muton_autopsy", Name: "Muton Autopsy", Cost: 60, AlienLore: true},
	{ID: "ethereal_autopsy", Name: "Ethereal Autopsy", Cost: 80, AlienLore: true, Requires: []string{"sectoid_autopsy", "floater_autopsy"}},

	// Weapons
	{ID: "laser_weapons", Name: "Laser Weapons", Cost: 120, Requires: []string{"alien_alloys"}, UnlockWeap: []string{"laser_pistol", "laser_rifle"}},
	{ID: "plasma_weapons", Name: "Plasma Weapons", Cost: 200, Requires: []string{"elerium", "sectoid_autopsy"}, UnlockWeap: []string{"plasma_pistol", "plasma_rifle"}},
	{ID: "heavy_plasma", Name: "Heavy Plasma", Cost: 250, Requires: []string{"plasma_weapons", "muton_autopsy"}, UnlockWeap: []string{"plasma_rifle"}},

	// Armour
	{ID: "personal_armour", Name: "Personal Armour", Cost: 80, Requires: []string{"alien_alloys"}, UnlockArmor: []string{"personal"}},
	{ID: "light_suit", Name: "Light Suit", Cost: 150, Requires: []string{"personal_armour", "alien_alloys"}, UnlockArmor: []string{"light"}},
	{ID: "medium_suit", Name: "Medium Suit", Cost: 200, Requires: []string{"light_suit"}, UnlockArmor: []string{"medium"}},
	{ID: "heavy_suit", Name: "Heavy Suit", Cost: 280, Requires: []string{"medium_suit"}, UnlockArmor: []string{"heavy"}},
	{ID: "power_suit", Name: "Power Suit", Cost: 400, Requires: []string{"heavy_suit", "elerium"}, UnlockArmor: []string{"power_suit"}},
	{ID: "flight_suit", Name: "Flying Suit", Cost: 500, Requires: []string{"power_suit"}, UnlockArmor: []string{"flight_suit"}},

	// Alien data
	{ID: "mind_control", Name: "Mind Control", Cost: 150, Requires: []string{"ethereal_autopsy"}, AlienLore: true},
	{ID: "ufo_nav", Name: "UFO Navigation", Cost: 100, AlienLore: true},
	{ID: "ufo_power", Name: "UFO Power Source", Cost: 120, AlienLore: true},
	{ID: "alien_comm", Name: "Alien Communications", Cost: 90, AlienLore: true},
}

func ResearchByID(id string) *ResearchTopic {
	for i := range ResearchTree {
		if ResearchTree[i].ID == id {
			return &ResearchTree[i]
		}
	}
	return nil
}
