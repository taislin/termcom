# termcom — Manual de la demake ASCII estilo roguelike de X-COM (v0.50.5)

## Tabla de contenidos

1. [Resumen](#resumen)
2. [Primeros pasos](#primeros-pasos)
3. [Tutorial / Incorporación](#tutorial--incorporación)
4. [Geoscape](#geoscape)
5. [Gestión de la base](#gestión-de-la-base)
6. [Investigación y fabricación](#investigación-y-fabricación)
7. [Equipar soldados](#equipar-soldados)
8. [Campo de Batalla](#campo-de-batalla)
9. [Armas y equipo](#armas-y-equipo)
10. [Armadura](#armadura)
11. [Alienígenas](#alienígenas)
12. [Rangos y progresión de soldados](#rangos-y-progresión-de-soldados)
13. [Guardar/Cargar](#guardarcargar)
14. [Referencia de teclas](#referencia-de-teclas)
15. [Consejos y estrategia](#consejos-y-estrategia)
16. [Tablas de referencia](#tablas-de-referencia)

---

## Resumen

**termcom** es una demake ASCII estilo roguelike de X-COM: UFO Defense (1994), renderizada
enteramente en la terminal. Comandáis a X-COM — una fuerza de tarea internacional que defiende
la Tierra de una invasión alienígena.

**Vuestro objetivo:** Investigar la tecnología alienígena, fabricar armas y armaduras,
y liderar escuadrones en combate táctico para eliminar la amenaza alienígena.

**Victoria:** Ganad suficientes batallas para desencadenar la misión final de Cydonia, y luego ganadla.

**Derrota:** La Actividad Alienígena alcanza el 100% — la invasión abruma la Tierra.

**Dificultad:** Elegid un nivel antes de empezar. Las dificultades mayores hacen a los alienígenas
más duros, los OVNIs más frecuentes y los fondos iniciales más ajustados.

- **Principiante** — Aliens más débiles, OVNIs más lentos, más fondos iniciales
- **Experimentado** — Estándar
- **Veterano** — Aliens más duros, OVNIs más rápidos, menos dinero
- **Genio** — Mucho más difícil en todos los aspectos
- **Superhumano** — Amenaza alienígena máxima

**Idioma:** Hay 8 idiomas disponibles — cambiadlos en la pantalla de Opciones.
Inglés, chino, español, francés, ruso, portugués, japonés, coreano.

**Opciones:** Pulsad `?` en cualquier pantalla para abrir la ayuda, o id a la pantalla de
Opciones para ajustar el resplandor, la iluminación, el sonido, el autoguardado, el temblor de
pantalla, el soporte de ratón, las líneas de cuadrícula, los cuadros de confirmación, el tema,
la velocidad de resolución, el volumen y el idioma.

---

## Tutorial / Incorporación

En vuestra primera partida (no se detectan archivos de guardado), aparece automáticamente un
**Briefing** del Comandante paso a paso tras seleccionar la dificultad. Cubre:

1. **Bienvenida** — Introducción a X-COM
2. **Geoscape y tiempo** — Pausa (`Space`), velocidad (`1`–`4`)
3. **Detección de OVNIs** — Radar y marcadores de OVNI
4. **Lanzamiento de interceptor** — Pulsad `L` para enfrentaros
5. **Respuesta a misión** — Pulsad `M` para desplegar
6. **Gestión de base** — Pulsad `B` para gestionar vuestra base
7. **Campo de Batalla** — Unidades de Tiempo, movimiento y combate
8. **Hecho** — Ya estáis listos

**Controles:** Enter avanza, S salta, Esc descarta.

**Repetir:** Abriendo la pantalla de Opciones y seleccionando "Repetir Tutorial" en cualquier momento.

---

## Primeros pasos

Empezáis en el **Geoscape** — el mapa mundial. El tiempo avanza automáticamente.
Los OVNIs aparecen en el radar según entran en rango.

**Recursos iniciales:**
- $500.000 (modificado por la dificultad)
- 10 científicos, 10 ingenieros
- Una base con Viviendas, Laboratorio, Taller, Almacén y Radar
- Varios rifles y pistolas

**Cuando se detecta un OVNI:**

| Acción | Tecla | Qué ocurre |
|--------|-------|------------|
| Lanzar interceptor | `L` | Enviar un caza para derribarlo |
| Enviar transporte | `R` | Enviar tropas a investigar un punto de impacto |
| Autoresolver | `A` | Resultado rápido de intercepción automática |
| Responder a misión | `M` | Desplegar en misiones de terror/suministros alienígenas |

---

## Geoscape

El Geoscape muestra un **panel regional** con los niveles de amenaza de cada región:

- **Panel izquierdo:** lista de regiones con barras de amenaza y estado del radar
- **Panel derecho:** minimapa ASCII que muestra bases, OVNIs, interceptores y rutas

### Símbolos del minimapa

| Símbolo | Significado |
|---------|-------------|
| ◆ | Vuestra base |
| ◉ | Nodo seleccionado actualmente |
| ○ | Centro regional (verde=seguro, amarillo=amenaza, rojo=peligro) |
| · | Anillo de cobertura del radar |
| ! | OVNI (rojo, negrita) |
| > | Interceptor patrullando |
| ► | Interceptor enfrentándose a un OVNI |
| ✕ | Interceptor u OVNI destruido |
| * | Punto de impacto (amarillo=sin saquear, gris=saqueado) |
| ≈ | Transporte en ruta (verde) |

### Controles de tiempo

| Tecla | Velocidad | Uso |
|-------|-----------|-----|
| Space | Pausa | Detener el tiempo para planear |
| 1 | 1x | Avance lento |
| 2 | 5x | Velocidad normal de patrulla |
| 3 | 20x | Avance rápido |
| 4 | 60x | Velocidad máxima |

### Presupuesto mensual

- **Ingresos:** $200.000 base + $50.000 por instalación de Radar
- **Gastos:** $2.000 por soldado, científico e ingeniero

### Múltiples bases

Pulsad `N` en un nodo vacío para construir una base nueva ($500K). Cada base tiene sus
propias instalaciones, soldados y almacenes. Pulsad `C` para ciclar la base activa.
Pulsad `T` para abrir la pantalla de Transferencia y mover soldados u objetos entre bases.

### Respuesta a misión

Cuando aparece una misión (terror, asalto a suministros, abducción, etc.), pulsad `M` para responder:

| Opción | Resultado |
|--------|-----------|
| **Desplegar escuadrón** | Combate táctico completo — mejores recompensas, mayor riesgo |
| **Autoresolver** | Resultado rápido — XP reducida, sin cadáveres, pero seguro |
| **Ignorar** | Saltárosla — la actividad alienígena sube |

Autoresolver da aproximadamente la mitad de la XP de una pelea real, sin cadáveres alienígenas,
y una pequeña probabilidad de bajas en caso de derrota.

### Defensa de base

Si una misión tiene como objetivo un nodo con vuestra base, responder lanza una batalla de
**Defensa de Base**. Perder una defensa de base destruye la base y su personal.
Perder vuestra última base termina la partida.

### Intercepción de OVNIs

Pulsad `L` para lanzar un interceptor contra el OVNI más cercano. El interceptor persigue
y se enfrenta en un breve combate aéreo autoresuelto. El minimapa muestra el enfrentamiento
con barras de HP y retroalimentación de acierto/fallo.

### Misiones alienígenas

Las misiones aparecen cada ~30 minutos de juego con un temporizador de 12–36 horas:

| Misión | Temporizador | Qué esperar |
|--------|--------------|-------------|
| Terror | 24h | Mapa urbano, muchos civiles en peligro |
| Asalto a suministros | 24h | Interior de OVNI, bonus de aleaciones/elerio |
| Abducción | 24h | Mapa rural, rescatar civiles |
| Investigación alienígena | 24h | Interior de OVNI, bonus de tecnología alienígena |
| Consejo | 36h | Mapa urbano, bonus de $100K en financiación |
| Asalto a base alienígena | 12h | Base alienígena rocosa, gran botín tecnológico |

Dejar que una misión expire aumenta la Actividad Alienígena en un 10%.

---

## Gestión de la base

Pulsad `B` desde el Geoscape para abrir vuestra base.

### Pestañas

| Tecla | Pestaña |
|-------|---------|
| 1 | Instalaciones |
| 2 | Soldados |
| 3 | Investigación |
| 4 | Fabricación |
| 5 | Transferencia |
| 6 | Hangares |

### Instalaciones

| Instalación | Coste | Tiempo de construcción | Efecto |
|-------------|-------|------------------------|--------|
| Viviendas | $50K | 5 días | +8 capacidad de soldados |
| Laboratorio | $75K | 7 días | Habilita la investigación |
| Taller | $60K | 7 días | Habilita la fabricación |
| Almacén | $40K | 3 días | +50 almacenamiento de objetos |
| Radar | $80K | 5 días | +$50K de financiación mensual |
| Contención alienígena | $100K | 10 días | Retiene hasta 10 alienígenas vivos |
| Laboratorio Psi | $150K | 14 días | Entrena habilidad psi |
| Hangar | $120K | 8 días | Aloja un interceptor |

**Bonus de adyacencia:** Colocar instalaciones del mismo tipo una junto a otra ayuda:
- Laboratorios adyacentes: investigación más rápida (hasta +30%)
- Talleres adyacentes: fabricación más rápida (hasta +30%)
- Viviendas adyacentes: curación de soldados más rápida (hasta +3 HP/día)

**Controles:** `B` para construir, `S` para vender (reembolso del 50%).

### Pestaña Soldados

Contratad soldados a $50K cada uno.

| Tecla | Acción |
|-------|--------|
| H | Contratar soldado |
| E | Abrir pantalla de equipo |
| G | Abrir diseñador de armas |
| D | Descartar soldado |

### Pestaña Hangares

Cada Hangar aloja un interceptor. Gestionad vuestra fuerza aérea aquí.

| Tecla | Acción |
|-------|--------|
| B | Comprar interceptor |
| W | Equipar arma |
| G | Abrir diseñador de armas |
| D | Abrir diseñador de aviones |

---

## Investigación y fabricación

### Investigación

Desde la pestaña de Investigación, asignad científicos a temas. La investigación avanza
automáticamente a medida que pasa el tiempo de juego.

El árbol tecnológico es **procedural** — cada partida genera un árbol único
a partir de un algoritmo con semilla. Las tecnologías centrales (Armas Láser, Armadura Personal,
Armas de Plasma) están siempre presentes, pero los prerrequisitos y costes varían.

**Prioridades:**
- **Aleaciones alienígenas** y **Elerio-115** se deben investigar primero
- Las autopsias de especies alienígenas desbloquean lore y pueden condicionar tecnologías de armas
- Interrogad a alienígenas capturados (tecla `I`) para completar la investigación más rápido

### Fabricación

Desde la pestaña de Fabricación, asignad ingenieros para producir objetos.
Esto produce **armas y armaduras de serie** — para equipo personalizado más potente,
usad el Diseñador de Armas y el Diseñador de Aviones.

**Objetos fabricables** (construid varios a la vez):

| Objeto | Tiempo | Materiales |
|--------|--------|------------|
| Pistola, Rifle, Cañón pesado, Cañón automático | 3–8 días | Aleaciones |
| Lanzacohetes, Vara aturdidora | 2–8 días | Aleaciones + Elerio |
| Armadura personal, Traje ligero/medio/pesado/de potencia/volador | 6–18 días | Aleaciones + Elerio |
| Medipack | 3 días | Aleaciones |

Más ingenieros = producción más rápida. Las armas de energía (Láser, Plasma) no se pueden
fabricar — deben investigarse, diseñarse y construirse mediante el
Diseñador de Armas, o recuperarse de alienígenas.

---

## Equipar soldados

Desde la pestaña de Soldados, pulsad `E` para abrir la pantalla de equipo.

### Controles

| Tecla | Acción |
|-------|--------|
| ↑/↓ | Seleccionar soldado |
| Tab | Ciclar objetos disponibles |
| 1 | Ranura de arma |
| 2 | Ranura de armadura |
| 3 | Ranura de inventario |
| Space | Equipar objeto seleccionado |
| G | Abrir diseñador de armas |
| A | Autoequipar todos los soldados |
| Esc | Volver |

### Ranuras

- **Ranura 1 (Arma):** Arma principal — diseñada a medida o un rifle/pistola de serie
- **Ranura 2 (Armadura):** Armadura corporal — personal, traje ligero, etc.
- **Ranura 3 (Inventario):** Objetos extra — granadas, medipacks, escáneres,
  minas de proximidad, amplificadores psi, armas cuerpo a cuerpo

### Sobrecarga

Cada objeto tiene peso. El peso total de arma + armadura + inventario
es vuestra **sobrecarga**. Una sobrecarga mayor reduce vuestras Unidades de Tiempo en batalla
(aprox. 1 UT de penalización por cada 5 unidades de peso). Mantened a vuestros soldados
cargados ligeramente para máxima movilidad.

### Autoequipar

Pulsad `A` para equipar automáticamente a cada soldado con el mejor arma
y armadura disponibles del almacén. El equipo existente vuelve al almacén — una forma
rápida de reequipar a vuestro escuadrón tras investigar nueva tecnología.

---

## Campo de Batalla

El Campo de Batalla es combate táctico por turnos. Controláis un escuadrón de soldados
contra fuerzas alienígenas en un mapa de cuadrícula de 50×50.

### Estructura de turnos

1. **Turno del jugador** — Moveos y actuad con cada soldado usando Unidades de Tiempo (UT)
2. **Turno alienígena** — Los alienígenas actúan usando sus propias reservas de UT
3. Repetid hasta que un bando es eliminado

### Unidades de Tiempo (UT)

Cada acción cuesta UT. Las UT se restauran por completo al inicio de cada turno del jugador.

| Acción | Coste aproximado de UT |
|--------|------------------------|
| Mover (por casilla) | 4 |
| Agacharse | 4 |
| Disparar arma | Varia según el arma (apuntado=base, ráfaga=1,5×, auto=2×) |
| Recargar | 8 |
| Lanzar granada | 20 |
| Usar medipack | 25 |
| Ataque psi | 20 |

La reserva de UT de un soldado empieza en 45–55 y puede crecer con la experiencia (máx ~80).

### Modos de fuego

Las armas pueden tener múltiples modos de fuego. Pulsad **Tab** para ciclar, y comprobad el
modo mostrado en la barra lateral.

| Modo | Coste | Precisión | Disparos | Cuándo usar |
|------|-------|-----------|----------|-------------|
| **Apuntado** | UT base | Mejor | 1 disparo | Larga distancia, objetivos de gran valor |
| **Ráfaga** | 1,5× UT | -10% | 3 disparos | Distancia media, supresión |
| **Auto** | 2× UT | -20% | Toda la munición restante | Corta distancia, emergencias |

No todas las armas soportan todos los modos. Los rifles y rifles láser soportan ráfaga;
solo unas pocas armas soportan fuego automático.

### Combate

Factores del combate:
- **Precisión** depende de la habilidad del soldado, la distancia al objetivo, la cobertura y el
  modo de fuego usado
- **Agacharse** da un bonus de precisión y reduce el daño recibido
- **Cobertura** — los muros bloquean un 80% del daño, los árboles 60%, los arbustos 40%,
  las vallas 30%. Colocad a vuestros soldados tras cobertura sólida
- **Esquivar la cobertura** con granadas — explosionan en un área e ignoran
  la reducción de daño por cobertura

### Línea de visión

Los soldados solo pueden ver en línea recta. Los muros, árboles y rocas bloquean la línea de visión (LOS).
Los suelos, puertas y la hierba no.

### Objetos y cobertura

Los disparos que pasan a través de objetos tienen el daño reducido por el valor de cobertura del objeto.
Se aplica la cobertura más alta a lo largo de la línea de fuego.

| Objeto | Cobertura % |
|--------|-------------|
| Muro / Muro OVNI | 80% |
| Roca | 70% |
| Árbol | 60% |
| Mobiliario OVNI | 50% |
| Arbusto | 40% |
| Humo denso | 40% |
| Valla | 30% |
| Escombros | 20% |

### Granadas

- Alcance: ~6 casillas
- Daño: basado en la fuerza, con salpicadura en área
- Destruye muros, árboles, rocas y vallas dentro del radio de la explosión
- Crea nubes de humo que bloquean la LOS a alta densidad

### Medipack

Cura 10 HP por uso (15 HP con la ventaja de Sanitario de Campo), cuesta 25 UT.

### Misiones nocturnas

Noche (antes de las 6:00 o después de las 18:00):
- Menor precisión (aprox. 75% de la diurna)
- Alcance de visión reducido (de 20 a ~10 casillas)
- Los soldados brillan tibiamente, los alienígenas brillan tenue azul

### Modos de visión

Pulsad `V` para ciclar: **Normal → Visión Nocturna → Térmica → Normal**
- Visión Nocturna: superposición de fósforo verde con estática
- Térmica: las entidades vivas brillan calientes, el terreno es azul frío

### Combate psi

Requiere un arma Amplificadora Psi y una instalación de Laboratorio Psi. El éxito depende de
la habilidad psi de vuestro soldado frente a la fuerza psi del objetivo. Un ataque psi
con éxito causa pánico en el objetivo — pierde su turno.

Los soldados en una base con Laboratorio Psi pueden ganar habilidad psi con el tiempo (hasta ~80).
La investigación de Control Mental otorga un gran impulso psi a todos los soldados.

### Modificadores de misión

Modificadores aleatorios que cambian en cada batalla:

| Modificador | Qué ocurre |
|-------------|------------|
| Ops nocturnas | Batalla nocturna forzada, botín bonus |
| Refuerzos | Llegan alienígenas extra en el turno 4 |
| Límite de tiempo | 15 turnos para eliminar a todos los alienígenas |
| Rescate de VIP | Proteger a un VIP, dinero bonus si sobrevive |
| Con trampas | Más granadas y minas en el mapa |
| Niebla espesa | Alcance de visión reducido un 40% |
| Emboscada alienígena | Los alienígenas empiezan en posiciones de vigilancia |
| Baja visibilidad | Precisión reducida para todas las unidades |
| Terreno elevado | Las posiciones elevadas dan bonus de precisión |

### Clima

El clima afecta al combate según la ubicación de la misión:

| Clima | Efecto |
|-------|--------|
| Lluvia | Menor precisión, el fuego se propaga más lento |
| Viento | El fuego se propaga más rápido, las granadas pueden derivar |
| Nieve | El movimiento cuesta más en nieve profunda |
| Niebla | Menor precisión, visión reducida |
| Tormenta | Lluvia + viento combinados |
| Frío | Ligera penalización de precisión |

### Informe de misión

Tras cada batalla, veis:
- Resultado (Victoria / Derrota)
- Alienígenas eliminados y soldados perdidos
- Botín recuperado y prisioneros capturados
- Fondos ganados
- Subidas de estadísticas por soldado o marca de "KIA"

Pulsad **Enter**, **Space** o **Esc** para descartar.

---

## Armas y equipo

### Diseñador de armas personalizado

Pulsad `G` desde la pantalla de Base, Soldados o Equipo para abrir el **Diseñador de Armas**.
Esta es la forma principal de crear armas para vuestros soldados. Elegid una plantilla base
y personalizad cada componente:

| Componente | Opciones | Qué afecta |
|------------|----------|------------|
| **Base** | Pistola / Rifle | Daño inicial, alcance, precisión, coste de UT |
| **Cañón** | Corto / Estándar / Largo / Extendido | Alcance, precisión, coste de UT, peso |
| **Óptica** | Ninguno / Miras de Hierro / Visor / Óptica Avanzada | Precisión, coste de UT, peso |
| **Modo de fuego** | Semi-Auto / Full-Auto | Modo automático (dispara más rápido, menos preciso) |
| **Munición** | Estándar / Perforante / Incendiaria / Explosiva | Modificador de daño, coste de UT, peso |
| **Culata** | Ninguno / Ligera / Pesada | Precisión, coste de UT, peso |

Cada componente afecta el daño, la precisión, el coste de UT, el alcance y el peso del arma.
El panel de vista previa muestra el arma ensamblada como arte ASCII coloreado y muestra
sus estadísticas finales. Los diseños se guardan como objetos personalizados disponibles en la pantalla de Equipo.

**Consejo:** Empezad con una base de Rifle para la mayoría de propósitos. Los cañones largos y los visores
mejoran la precisión a distancia. La munición explosiva golpea fuerte pero cuesta UT extra.

### Armas de serie

Estos objetos base están disponibles desde el inicio y se pueden fabricar:

| Tipo | Daño | Munición | Notas |
|------|------|----------|-------|
| Pistola | Ligero | Balística | Necesita recarga, poco peso |
| Rifle | Medio | Balística | Equipo estándar, soporta ráfaga |
| Cañón pesado | Alto | Balística | Lento, pesado, golpea fuerte |
| Cañón automático | Medio | Balística | Opción de fuego automático |
| Lanzacohetes | Muy alto | Explosiva | Daño en área |

### Armas de energía

Investigadas más tarde — nunca necesitan recarga:

| Tipo | Daño | Notas |
|------|------|-------|
| Pistola láser | Ligero | Arma de energía temprana |
| Rifle láser | Medio | Soporta fuego de ráfaga, nunca recarga |
| Pistola de plasma | Medio | Arma alienígena, nunca recarga |
| Rifle de plasma | Alto | Arma alienígena, nunca recarga |
| Plasma pesado | Muy alto | Arma alienígena de gama alta |

### Munición y recarga

- **Armas balísticas** necesitan recarga — pulsad `R` en combate
- **Armas de energía** (Láser, Plasma) nunca necesitan recarga
- **Consumibles** (granadas, medipacks) se usan desde vuestro inventario

### Modos de fuego

Ver [Campo de Batalla → Modos de fuego](#modos-de-fuego) para más detalles.

### Objetos de inventario

Los soldados pueden llevar objetos extra en su ranura de inventario:
- **Granadas** — explosivo arrojadizo con daño en área
- **Medipacks** — curaros a vosotros mismos o a un aliado adyacente
- **Escáneres de movimiento** — detectan enemigos cercanos
- **Minas de proximidad** — colocadas en el suelo, detonan cuando un enemigo pasa por encima
- **Amplificadores Psi** — habilitan ataques psi (requiere habilidad psi)
- **Armas cuerpo a cuerpo** — Vara aturdidora para derribos no letales

Cada objeto de inventario añade peso y aumenta la sobrecarga, reduciendo vuestras
UT disponibles en batalla. Empaquetad con criterio.

### Objetos procedurales

Cada partida genera armas y armaduras únicas basadas en la especie alienígena
encontrada. Tienen nombres y estadísticas aleatorias — cada juego es diferente.

**Armas procedurales:** 2–3 armas con tipos de daño que coinciden con la especie alienígena.
**Armaduras procedurales:** 1–2 piezas de armadura con protección que coincide con los tipos de daño alienígena.

Estos objetos se añaden automáticamente a vuestro almacén al inicio de la partida.

---

## Armadura

| Armadura | Defensa | Penalización UT | Notas |
|----------|---------|-----------------|-------|
| Ninguna | 0 | Ninguna | Por defecto |
| Armadura personal | 10 | Ninguna | Estándar de principios de partida |
| Traje ligero | 20 | -5% UT | Buena opción de mitad de partida |
| Traje medio | 30 | -10% UT | Protección fuerte |
| Traje pesado | 40 | -15% UT | Defensa máxima, penalización pesada |
| Traje de potencia | 50 | -10% UT | Armadura de final de partida |
| Traje volador | 45 | -5% UT | Casi final de partida, más ligero que el de potencia |

Una defensa mayor reduce el daño recibido, pero los trajes más pesados cuestan Unidades de Tiempo.

---

## Alienígenas

### Especies procedurales

Cada juego genera de 5 a 7 especies alienígenas únicas a partir de una semilla. Cada especie tiene
2–5 variantes de rango (Soldado → Navegante → Comandante → Élite → Soberano).

Las especies difieren en:
- **Tipo de daño** — la clase de daño que infligen
- **Resistencias y debilidades** — algunos son débiles al plasma, otros a los explosivos
- **Preferencia de arma** — los rangos bajos usan pistolas, los altos usan armas pesadas
- **Morfología** — plan corporal físico que afecta estadísticas y resistencias

Esto significa que **cada partida presenta amenazas alienígenas diferentes**. Una partida puede
tener una especie con fuerte componente psiónico, otra puede tener depredadores cuerpo a cuerpo débiles a los explosivos.

### Morfología

La morfología determina la forma física de un alienígena. Factores clave:

**Extremidades:**
- Brazos (0–6): Menos brazos = peor precisión, más brazos = mejor estabilidad o doble empuñadura
- Piernas (0–8): Más piernas = más rápido pero objetivo mayor; cero piernas = flotante, más difícil de acertar

**Tipos de cuerpo y sus resistencias:**
- **Carne de carbono:** +Resistencia cinética, -Debilidad explosiva
- **Basado en silicio:** +Resistencia Láser/+Plasma, -Debilidad explosiva, reflectante
- **Gaseoso:** Inmune a cinético, débil a plasma, puede atravesar muros en fase
- **Cristalino:** Buena resistencia general, muy débil a explosivos, se hace añicos al morir
- **Amorfo:** +Resistencia psi, regenera HP cada turno
- **Mecánico:** Inmune a psi, +Resistencia Plasma, -Debilidad láser, autodestrucción
- **Bio-Sintético:** Resistencias equilibradas, cura alienígenas adyacentes
- **Nanotecnológico:** +Resistencia cinética, puede revivir al morir

**Sentidos:**
- **Vista:** Afecta a la precisión — el espectro múltiple ignora humo/oscuridad
- **Oído:** La ecolocalización detecta unidades a través del humo a corta distancia
- **Sentido térmico:** Detecta unidades vivas sin importar la cobertura a corta distancia
- **Sentido psiónico:** Potencia psi, detecta humanos controlados mentalmente
- **Sentido químico:** Bonus de precisión contra objetivos heridos

### Niveles de conocimiento

A medida que encontráis alienígenas, mejora la inteligencia:

| Nivel | Qué aprendéis |
|-------|---------------|
| Desconocido | El nombre aparece como "???" |
| Avistado | Nombre e icono revelados |
| Eliminado | Estadísticas y resistencias reveladas |
| Autopsiado | Lore completo y debilidades detalladas |

### IA alienígena

Los alienígenas patrullan hasta que detectan a un humano, entonces atacan. Comportamientos incluyen:
- **Buscar** — moverse hacia la última posición conocida durante unos turnos
- **Huir** — alejarse cuando están muy heridos y con baja valentía
- **Adaptar** — los alienígenas estudian vuestras tácticas entre misiones.
  ¿Francotiradores a distancia? Cargarán contra vosotros. ¿Usáis granadas? Se dispersarán.
  ¿Flanqueáis a menudo? Colocarán supresores.

### Escalada de equipo

Los alienígenas obtienen mejor equipo a medida que avanza la campaña:
- **Meses iniciales:** Pistolas de plasma, armadura básica
- **Mitad de campaña:** Rifles de plasma, plasma pesado, cañones alienígenas
- **Final de campaña:** Armas y armaduras alienígenas de gama alta

### Captura alienígena

Usad una **Vara aturdidora** (cuerpo a cuerpo, $2K para fabricar) para dejar alienígenas inconscientes.
Si el daño de aturdimiento supera su HP, caen inconscientes y se pueden recoger
tras la misión — siempre que tengáis Contención Alienígena con capacidad libre.

Los alienígenas capturados se pueden interrogar desde la pantalla de Investigación (tecla `I`):
- El interrogatorio puede completar una autopsia activa al instante
- O conceder un bonus de progreso a la investigación actual
- Requiere al menos un Laboratorio

---

## Rangos y progresión de soldados

### Rangos

Los rangos se desbloquean a medida que crece vuestra plantilla total:

| Rango | Se desbloquea cuando la plantilla alcanza |
|-------|--------------------------------------------|
| Recluta | Siempre disponible |
| Cabo | Siempre disponible |
| Cabo primero | 4 soldados |
| Sargento | 8 soldados |
| Teniente | 14 soldados |
| Capitán | 22 soldados |
| Comandante | 30 soldados |
| Coronel | 40 soldados |

### Crecimiento de estadísticas

Los soldados mejoran mediante **experiencia por acción** durante la batalla:
- **Disparo** → mejora la Precisión
- **Reacciones** → mejora las Reacciones
- **Cuerpo a cuerpo** → mejora la Fuerza
- **Valentía** → mejora la Valentía (al resistir el pánico)
- **Habilidad psi** → mejora la Habilidad Psi y la Fuerza Psi

Tras cada misión, la XP acumulada se convierte en subidas de estadísticas. Los soldados que
ganaron XP también obtienen crecimiento general de "halo" hacia sus límites de HP, UT y Fuerza.
Los límites son aproximadamente: UT 80, HP 60, Precisión 120, Reacciones 100,
Valentía 100, Fuerza 70, Psi 100.

### Fatiga y heridas

- **Soldados heridos** no pueden desplegarse hasta curarse (2 HP/día de recuperación)
- **Fatiga:** Las batallas causan 1–5 días de fatiga
- Las instalaciones de curación y las Viviendas aceleran la recuperación

### Heridas mortales

En batalla, los impactos pueden causar heridas mortales y sangrado. El sangrado drena HP cada
turno — ponedles un medipack rápido. Las heridas que sobreviven se convierten en días de
recuperación tras la misión.

### Moral

Los soldados recuperan moral cada turno. La baja moral puede desencadenar pánico (saltar turno).
Resistir el pánico construye XP de valentía.

### Ventajas

Cada subida de rango otorga una ventaja aleatoria:

| Ventaja | Efecto |
|---------|--------|
| Reflejos rápidos | +10 Reacciones |
| Francotirador | +Precisión a larga distancia |
| Granadero | Mayor salpicadura de granada |
| Sanitario de campo | El medipack cura más |
| Voluntad de hierro | +Habilidad Psi y +Fuerza Psi |
| Puntería firme | +Precisión al estar quieto |
| Especialista en combate cercano | +Precisión a corta distancia |
| Experto en vigilancia | +Precisión de fuego de reacción |
| Demoliciones | +Daño de granada |
| Buscador | +Botín de batallas |
| Duro | +5 HP máximo |
| Aprendiz rápido | +Ganancia de XP |

### Memorial

Los soldados caídos en acción se registran en el Memorial del juego.
Podéis verlo para honrar a los caídos.

---

## Guardar/Cargar

| Tecla | Acción |
|-------|--------|
| F5 | Abrir selector de ranura de guardado |
| F9 | Abrir selector de ranura de carga |

Los guardados incluyen: tiempo de juego, fondos, estado de pausa, actividad alienígena, estado de la base,
OVNIs, misiones activas, semilla de especies procedurales y niveles de conocimiento alienígena.
La semilla asegura que las mismas especies alienígenas se regeneren al cargar.

**Autoguardado:** Si está habilitado en Opciones, el juego autoguarda periódicamente.

---

## Referencia de teclas

### Geoscape

| Tecla | Acción |
|-------|--------|
| Teclas de flecha | Mover cámara |
| j/k | Navegar lista de regiones |
| Space | Pausa/reanuda |
| 1–4 | Velocidad del tiempo |
| B | Abrir base |
| L | Lanzar interceptor |
| A | Autoresolver OVNI más cercano |
| M | Responder a misión |
| R | Enviar transporte a punto de impacto |
| C | Ciclar a la siguiente base |
| N | Construir base nueva ($500K) |
| T | Abrir pantalla de transferencia |
| E | Abrir enciclopedia |
| V | Alternar superposición de radar |
| F5 | Guardar |
| F9 | Cargar |
| Q | Salir |
| ? | Ayuda |

### Gestión de base

| Tecla | Acción |
|-------|--------|
| 1–6 | Cambiar pestaña |
| j/k | Navegar objetos |
| B | Construir instalación |
| S | Vender instalación |
| H | Contratar soldado |
| E | Abrir pantalla de equipo |
| G | Abrir diseñador de armas |
| D | Descartar soldado / Diseñador de aviones (Hangares) |
| Esc | Volver al geoscape |

### Pantalla de equipo

| Tecla | Acción |
|-------|--------|
| ↑/↓ | Seleccionar soldado |
| Tab | Ciclar objetos disponibles |
| 1 | Ranura de arma |
| 2 | Ranura de armadura |
| 3 | Ranura de inventario |
| Space | Equipar objeto seleccionado |
| G | Abrir diseñador de armas |
| A | Autoequipar todos los soldados |
| Esc | Volver |

### Diseñador de armas

Pulsad `G` desde Base, Soldados o Equipo.

| Parámetro | Opciones | Efecto |
|-----------|----------|--------|
| Cañón | Corto / Estándar / Largo / Extendido | Alcance, precisión, peso, coste UT |
| Óptica | Ninguno / Miras de Hierro / Visor / Avanzada | Precisión, peso, coste UT |
| Modo de fuego | Semi-Auto / Full-Auto | Modo automático |
| Munición | Estándar / Perforante / Incendiaria / Explosiva | Daño, peso, coste UT |
| Culata | Ninguno / Ligera / Pesada | Precisión, peso, coste UT |

### Diseñador de aviones (Interceptores personalizados)

Todos los interceptores se diseñan y construyen mediante el **Diseñador de Aviones**.
Pulsad `D` desde la pestaña de Hangares para abrirlo. Configurad vuestra aeronave:

| Parámetro | Rango | Qué afecta |
|-----------|-------|------------|
| **Longitud** | Corto (3) → Largo (7) | Puntos de casco, masa, velocidad |
| **Envergadura** | Corto (1) → Ancho (4) | Maniobrabilidad, masa |
| **Motores** | 1–3 | Velocidad, capacidad de combustible, masa |
| **Combustible** | 20–100 | Alcance operativo |
| **Arma** | Cañón / Stingray / Avalancha / Plasma | Fuego, peso, coste |
| **Blindaje** | Ninguno / Aleación ligera / Aleación pesada / Revestimiento alienígena | Bonus de casco, reducción de daño, masa |

El diseñador calcula estadísticas derivadas (velocidad, fuego, casco, relación masa/empuje)
a partir de vuestra configuración y muestra una vista previa ASCII coloreada. Los diseños
más pesados son más resistentes pero más lentos — equilibrad durabilidad frente a velocidad de intercepción.

**Armas de avión:**

| Arma | Daño | Precisión | Alcance | Cadencia | Coste |
|------|------|-----------|---------|----------|-------|
| Cañón | 15 | 85% | 25 | 3 disparos | $5K |
| Stingray | 25 | 70% | 45 | 2 disparos | $8K |
| Avalancha | 40 | 55% | 60 | 1 disparo | $12K |
| Plasma | 60 | 50% | 50 | 1 disparo | $20K |

**Blindaje de avión:**

| Blindaje | Bonus de casco | Reducción de daño | Coste |
|----------|----------------|-------------------|-------|
| Ninguno | 0 | 0% | Gratis |
| Aleación ligera | +10 | 10% | $8K |
| Aleación pesada | +25 | 25% | $18K |
| Revestimiento alienígena | +40 | 40% | $35K |

### Campo de Batalla

| Tecla | Acción |
|-------|--------|
| Teclas de flecha / WASD / hjkl | Mover cursor |
| Space / Enter | Seleccionar unidad / confirmar |
| q | Ciclar soldados |
| f | Disparar arma |
| Tab | Ciclar modo de fuego |
| r | Recargar |
| e / n | Terminar turno |
| c | Agacharse |
| g | Lanzar granada |
| m | Modo mover |
| h | Usar medipack |
| p | Ataque psi |
| y | Escáner de movimiento |
| t | Colocar mina de proximidad |
| v | Ciclar modo de visión |
| o | Opciones |
| ? | Ayuda |
| Esc | Cancelar / deseleccionar |

### Controles táctiles móviles

En el navegador con pantalla estrecha (cols < 100) o cuando `touch_mode` está habilitado:

| Gesto | Acción |
|-------|--------|
| Tocar | Seleccionar, mover, disparar |
| Pulsación larga (500ms) | Cancelar |
| Arrastrar vertical | Desplazar |

Un botón `[=]` abre un menú de control en pantalla amigable al tacto.

---

## Consejos y estrategia

### Principios de partida

1. Investigad **Aleaciones alienígenas** primero — desbloquea Armas Láser y Armadura.
2. Construid un segundo **Radar** — más detección, más financiación mensual.
3. Contratad 2–4 soldados extra para llenar vuestros escuadrones.
4. Usad el **Diseñador de Armas** para crear rifles personalizados — podéis construir
   mejores armas que los modelos de serie con los componentes adecuados.
5. No ignoréis las autopsias — algunas tecnologías de armas las requieren.
6. Diseñad a medida vuestros interceptores en el **Diseñador de Aviones** — un diseño
   equilibrado (longitud/envergadura medias, 2 motores, misiles Stingray) supera
   a los interceptores por defecto.

### Combate

- **Usad cobertura** — muros (80%) > rocas (70%) > árboles (60%) > arbustos (40%)
- **Agachaos** antes de disparar para mejor precisión y reducción de daño
- **Las granadas** esquivan la cobertura y destruyen muros — perfectas para enemigos atrincherados
- **Aprended resistencias alienígenas** — consultad la enciclopedia tras las primeras bajas
- **No os extendáis demasiado** — los alienígenas disparan de reacción cuando os movéis en su LOS
- **Manted un médico** — un soldado con medipack puede salvar vidas
- **Gestionad la sobrecarga** — no sobrecarguéis a los soldados con equipo pesado

### Economía

- Vended cadáveres y botín alienígena sobrante por dinero
- Los sueldos mensuales se acumulan — equilibrad vuestra plantilla frente a los ingresos
- Las misiones del Consejo pagan bonus de $100K — priorizadlas
- Fabricad objetos para vender con beneficio a principios de partida

### Ruta de investigación

Aleaciones → Armas Láser → Armadura Personal → Autopsias → Elerio → Armas de Plasma

Mitad de partida: Traje medio, Plasma pesado.
Final de partida: Traje de potencia/volador, Control Mental.

### Construcción de base

- Los radares se pagan a sí mismos (+$50K/mes cada uno)
- Construid Almacén pronto — os llenaréis rápido
- La Contención Alienígena es necesaria para capturas en vivo y bonus de interrogatorio
- Las instalaciones adyacentes se potencian entre sí — planificad el diseño de vuestra base
- Construid un Laboratorio Psi si queréis capacidades psi

---

## Tablas de referencia

### Tipos de casilla

| Tipo | Carácter | Cobertura | Notas |
|------|----------|-----------|-------|
| Suelo | `.` | 0% | Terreno por defecto |
| Muro | `#` | 80% | Bloquea movimiento y LOS |
| Puerta | `+` | 0% | Se abre al contacto |
| Ventana | `¤` | 0% | Bloquea LOS, transitable |
| Hierba | `·` | 0% | Inflamable |
| Árbol | `♣` | 60% | Bloquea LOS, inflamable |
| Roca | `∩` | 70% | Bloquea LOS |
| Agua | `≈` | 0% | Intransitable |
| Suelo OVNI | `≡` | 0% | Suelo interior |
| Muro OVNI | `█` | 80% | Bloquea movimiento y LOS |
| Escaleras | `▓` / `▒` | 0% | Transición de nivel |
| Pavimento | `░` | 0% | Carretera / zona de aterrizaje |
| Arena | `·` | 0% | Terreno desértico |
| Nieve | `∗` | 0% | Terreno polar |
| Pantano | `≋` | 0% | Terreno pantanoso |
| Arbusto | `†` | 40% | Cobertura ligera, inflamable |
| Valla | `║` | 30% | Cobertura mínima, inflamable |
| Escombros | `▒` | 20% | Terreno destruido |
| Mobiliario | `■` `⚙` `◈` `⌁` `▤` `⊕` `⎔` `⊟` `⌨` `□` `◫` `⊞` | 50% | Objetos de OVNI/edificio |

### Tipos de daño

| Tipo | Fuente |
|------|--------|
| Cinético | Pistola, Rifle, Cañón pesado, Cañón automático |
| Láser | Pistola láser, Rifle láser |
| Plasma | Pistola de plasma, Rifle de plasma, Plasma pesado, Granada alienígena |
| Explosivo | Lanzacohetes, Granadas, Minas |
| Cuerpo a cuerpo | Vara aturdidora, Garra Chryssalid, Garra Reaper |
| Psiónico | Ataques psi etéreos |

### Terreno destructible

| Casilla | Se convierte en | Cambio de cobertura |
|---------|-----------------|---------------------|
| Muro / Muro OVNI | Escombros | 80% → 20% |
| Árbol | Escombros | 60% → 20% |
| Roca | Escombros | 70% → 20% |
| Valla | Escombros | 30% → 20% |

### Gas (de granadas)

| Densidad | Visual | ¿Bloquea LOS? | Penalización de cobertura |
|----------|--------|---------------|---------------------------|
| Denso (3) | ▓ | Sí | 40% |
| Medio (2) | ▒ | No | 20% |
| Tenue (1) | ░ | No | 0% |

El gas se extiende a casillas adyacentes y se adelgaza cada turno hasta disiparse.

### Fuego

| Propiedad | Detalle |
|-----------|---------|
| Animación | Cicla `^` → `w` → `*` |
| Colores | Amarillo → Naranja → Rojo |
| Propagación | Probabilidad por turno a casilla inflamable adyacente |
| Inflamable | Hierba, Árbol, Arbusto, Valla, Puerta |
| Duración | ~3 turnos, luego la casilla se convierte en Suelo |

### Fabricación

| Objeto | Tiempo | Materiales |
|--------|--------|------------|
| Pistola | 3 días | 1 aleación |
| Rifle | 5 días | 2 aleaciones |
| Cañón pesado | 7 días | 3 aleaciones |
| Cañón automático | 6 días | 3 aleaciones |
| Lanzacohetes | 8 días | 4 aleaciones, 1 elerio |
| Vara aturdidora | 2 días | 1 aleación |
| Armadura personal | 6 días | 2 aleaciones |
| Traje ligero | 10 días | 4 aleaciones, 1 elerio |
| Traje medio | 14 días | 6 aleaciones, 2 elerio |
| Traje pesado | 18 días | 8 aleaciones, 3 elerio |
| Medipack | 3 días | 1 aleación |
