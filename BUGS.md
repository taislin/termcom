# termcom Bug and TODO list

- [ ] implement different move speeds (sand slower, marsh slower, water slower etc, pavement slighly faster than grass, mud same as grass if dry, slower if wet, add this variation and make rainy weather make it muddy)

Add these as well and make the maps generate them:

- [ ] Shootable Streetlamps / Floodlights
Visuals: A tall pole ⇡ emitting a bright 7x7 radius of yellow/white background color, overriding the Fog of War.
Tactical Value: If your soldiers are pinned down in the light, they have no concealment. You can target and shoot the light source ⇡. The light goes out, plunging the area into pitch black tcell.ColorBlack, instantly giving your squad full concealment (unless the alien has Thermal vision!).

- [ ] Broken Glass / Debris
Visuals: Scattered commas , or backticks ` on the floor.
Tactical Value: Costs the normal amount of Time Units (TU) to walk through, but it makes a loud crunching noise. The engine registers an "Audio Event," and any alien AI within 15 tiles will immediately turn and walk toward that coordinate.

- [ ] Cryo-Coolant Pipes
Visuals: A thick metal pipe ═ carrying blue liquid.
Tactical Value: If shot, it doesn't explode. Instead, it violently vents freezing gas. Any tile within a 3x3 radius permanently becomes "Frozen" ≈ (Cyan). Any unit caught in the blast loses all their TU for that turn (Frozen Solid).

- [ ] Collapsible Floors / Glass Skylights
Visuals: A translucent floor tile ▒ on Level 1 (the roof).
Tactical Value: If a heavy unit (like a Muton or a soldier in Power Armor) walks on it, or if it takes damage, the floor shatters. The unit physically falls to Level 0, taking fall damage, and creating a permanent hole in the ceiling that can be shot through.
