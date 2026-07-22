# termcom — Manual do Demake Roguelike em ASCII X-COM (v0.51.18)

## Índice

1. [Visão Geral](#visão-geral)
2. [Primeiros Passos](#primeiros-passos)
3. [Tutorial / Integração](#tutorial--integração)
4. [Geomapa](#geomapa)
5. [Gestão da Base](#gestão-da-base)
6. [Investigação e Fabrico](#investigação-e-fabrico)
7. [Equipar Soldados](#equipar-soldados)
8. [Campo de Batalha](#campo-de-batalha)
9. [Armas e Equipamento](#armas-e-equipamento)
10. [Armadura](#armadura)
11. [Aliens](#aliens)
12. [Postos e Progressão dos Soldados](#postos-e-progressão-dos-soldados)
13. [Guardar/Carregar](#guardarcarregar)
14. [Referência de Teclas](#referência-de-teclas)
15. [Dicas e Estratégia](#dicas-e-estratégia)

---

## Visão Geral

**termcom** é um demake roguelike em ASCII de X-COM: UFO Defense (1994), renderizado
inteiramente num terminal. Você comanda o X-COM — uma força-tarefa internacional que
defende a Terra de uma invasão alienígena.

**O seu objetivo:** Investigar tecnologia alienígena, fabricar armas e armaduras,
e liderar esquadrões em combate tático para eliminar a ameaça alienígena.

**Vitória:** Vença batalhas suficientes para desencadear a missão final de Cydonia,
depois vença-a.

**Derrota:** A Atividade Alienígena atinge 100% — a invasão submerge a Terra.

**Dificuldade:** Escolha um nível antes de começar. Dificuldades mais altas tornam os
aliens mais resistentes, os OVNIs mais frequentes e os fundos iniciais mais escassos.

- **Iniciante** — Aliens mais fracos, OVNIs mais lentos, mais fundos iniciais
- **Experiente** — Padrão
- **Veterano** — Aliens mais fortes, OVNIs mais rápidos, menos dinheiro
- **Génio** — Muito mais difícil em geral
- **Super-humano** — Ameaça alienígena máxima

**Idioma:** 8 idiomas disponíveis — alterne no ecrã de Opções.
Inglês, Chinês, Espanhol, Francês, Russo, Português, Japonês, Coreano.

**Opções:** Prima `?` em qualquer ecrã para abrir a ajuda, ou navegue até ao ecrã
de Opções para ajustar o brilho, iluminação, som, gravação automática, tremor de
ecrã, suporte para rato, linhas de grelha, caixas de confirmação, tema, velocidade de
resolução, volume e idioma.

---

## Tutorial / Integração

Na sua primeira partida (sem ficheiros de gravação detetados), um **Briefing** passo
a passo do Comandante aparece automaticamente após selecionar a dificuldade. Cobre:

1. **Boas-vindas** — Introdução ao X-COM
2. **Geomapa e Tempo** — Pausa (`Space`), velocidade (`1`–`4`)
3. **Deteção de OVNIs** — Radar e marcadores de OVNIs
4. **Lançamento de Interceptor** — Prima `L` para engajar
5. **Resposta a Missão** — Prima `M` para destacar
6. **Gestão da Base** — Prima `B` para gerir a sua base
7. **Campo de Batalha** — Unidades de Tempo, movimento e combate
8. **Concluído** — Está pronto

**Controlos:** Enter avança, S ignora, Esc descarta.

**Repetir:** Abra o ecrã de Opções e selecione "Repetir Tutorial" a qualquer momento.

---

## Primeiros Passos

Você começa no **Geomapa** — o mapa-múndi. O tempo avança automaticamente.
Os OVNIs aparecem no radar à medida que entram em alcance.

**Recursos iniciais:**
- $500.000 (modificado pela dificuldade)
- 10 cientistas, 10 engenheiros
- Uma base com Quartos de Habitação, Laboratório, Oficina, Armazém e Radar
- Várias espingardas e pistolas

**Quando um OVNI é detetado:**

| Ação | Tecla | O que acontece |
|--------|-----|-------------|
| Lançar interceptor | `L` | Enviar um caça para o abater |
| Enviar transporte | `R` | Enviar tropas para investigar um local de acidente |
| Autoresolver | `A` | Resultado rápido de interceção automática |
| Responder a missão | `M` | Destacar para missões de terror/abastecimento alienígenas |

---

## Geomapa

O Geomapa mostra um **painel regional** com níveis de ameaça para cada região:

- **Painel esquerdo:** lista de regiões com barras de ameaça e estado do radar
- **Painel direito:** minimapa ASCII mostrando bases, OVNIs, interceptores e rotas

### Símbolos do Minimapa

| Símbolo | Significado |
|--------|---------|
| ◆ | A sua base |
| ◉ | Nó atualmente selecionado |
| ○ | Nó regional (verde=seguro, amarelo=ameaça, vermelho=perigo) |
| · | Anel de cobertura do radar |
| ! | OVNI (vermelho, negrito) |
| > | Interceptor em patrulha |
| ► | Interceptor a engajar um OVNI |
| ✕ | Interceptor ou OVNI destruído |
| * | Local de acidente (amarelo=por saquear, cinzento=saqueado) |
| ≈ | Transporte em rota (verde) |

### Controlos de Tempo

| Tecla | Velocidade | Uso |
|-----|-------|-----|
| Space | Pausa | Parar o tempo para planear |
| 1 | 1x | Avanço lento |
| 2 | 5x | Velocidade normal de patrulha |
| 3 | 20x | Avanço rápido |
| 4 | 60x | Velocidade máxima |

### Orçamento Mensal

- **Receita:** $200.000 base + $50.000 por instalação de Radar
- **Despesas:** $2.000 por soldado, cientista e engenheiro

### Múltiplas Bases

Prima `N` num nó vazio para construir uma nova base ($500K). Cada base tem as suas
próprias instalações, soldados e armazéns. Prima `C` para circular a base ativa.
Prima `T` para abrir o ecrã de Transferência e mover soldados ou itens entre bases.

### Resposta a Missão

Quando aparece uma missão (terror, incursão de abastecimento, sequestro, etc.),
prima `M` para responder:

| Opção | Resultado |
|--------|--------|
| **Destacar esquadrão** | Combate tático completo — melhores recompensas, maior risco |
| **Autoresolver** | Resultado rápido — XP reduzido, sem cadáveres, mas seguro |
| **Ignorar** | Saltar — a atividade alienígena sobe |

Autoresolver dá cerca de metade do XP de um combate real, sem cadáveres alienígenas,
e uma pequena hipótese de baixas em caso de derrota.

### Defesa da Base

Se uma missão tiver como alvo um nó com a sua base, responder inicia uma batalha de
**Defesa da Base**. Perder uma defesa de base destrói a base e o seu pessoal.
Perder a sua última base termina o jogo.

### Interceção de OVNIs

Prima `L` para lançar um interceptor contra o OVNI mais próximo. O interceptor
persegue e engaja num curto combate aéreo autoresolvido. O minimapa mostra o
engajamento com barras de HP e feedback de acerto/falha.

### Missões Aliens

As missões aparecem a cada ~30 minutos de jogo com um temporizador de 12–36 horas:

| Missão | Temporizador | O que esperar |
|---------|-------|----------------|
| Terror | 24h | Mapa urbano, muitos civis em perigo |
| Incursão de Abastecimento | 24h | Interior do OVNI, bónus de ligas/elério |
| Sequestro | 24h | Mapa rural, resgatar civis |
| Investigação Aliens | 24h | Interior do OVNI, bónus de tecnologia alienígena |
| Conselho | 36h | Mapa urbano, bónus de $100K em financiamento |
| Assalto a Base Aliens | 12h | Base alienígena rochosa, grande ganho de tecnologia |

Deixar expirar uma missão aumenta a Atividade Aliens em 10%.

---

## Gestão da Base

Prima `B` a partir do Geomapa para abrir a sua base.

### Abas

| Tecla | Aba |
|-----|-----|
| 1 | Instalações |
| 2 | Soldados |
| 3 | Investigação |
| 4 | Fabrico |
| 5 | Transferência |
| 6 | Hangares |

### Instalações

| Instalação | Custo | Tempo de Construção | Efeito |
|----------|------|------------|--------|
| Quartos de Habitação | $50K | 5 dias | +8 capacidade de soldados |
| Laboratório | $75K | 7 dias | Permite investigação |
| Oficina | $60K | 7 dias | Permite fabrico |
| Armazém | $40K | 3 dias | +50 armazenamento de itens |
| Radar | $80K | 5 dias | +$50K de financiamento mensal |
| Contenção de Aliens | $100K | 10 dias | Mantém até 10 aliens vivos |
| Lab Psi | $150K | 14 dias | Treina capacidade psi |
| Hangar | $120K | 8 dias | Abriga um interceptor |

**Bónus de adjacência:** Colocar instalações do mesmo tipo lado a lado ajuda:
- Laboratórios adjacentes: investigação mais rápida (até +30%)
- Oficinas adjacentes: fabrico mais rápido (até +30%)
- Quartos de Habitação adjacentes: cura de soldados mais rápida (até +3 HP/dia)

**Controlos:** `B` para construir, `S` para vender (reembolso de 50%).

### Aba de Soldados

Contrate soldados a $50K cada.

| Tecla | Ação |
|-----|--------|
| H | Contratar soldado |
| E | Abrir ecrã de equipamento |
| G | Abrir projetor de armas |
| D | Dispensar soldado |

### Aba de Hangares

Cada Hangar abriga um interceptor. Gira a sua força aérea aqui.

| Tecla | Ação |
|-----|--------|
| B | Comprar interceptor |
| W | Equipar arma |
| G | Abrir projetor de armas |
| D | Abrir projetor de aviões |

---

## Investigação e Fabrico

### Investigação

A partir da aba de Investigação, atribua cientistas a tópicos. A investigação
progressa automaticamente à medida que o tempo de jogo passa.

A árvore tecnológica é **procedural** — cada partida gera uma árvore única a partir
de um algoritmo com semente. As tecnologias centrais (Armas Laser, Armadura Pessoal,
Armas de Plasma) estão sempre presentes, mas pré-requisitos e custos variam.

**Prioridades:**
- **Ligas Aliens** e **Elério-115** devem ser investigados primeiro
- Autópsias de espécies alienígenas desbloqueiam lore e podem condicionar tecnologias de armas
- Interrogue aliens capturados (tecla `I`) para concluir investigações mais rápido

### Fabrico

A partir da aba de Fabrico, atribua engenheiros para produzir itens.
Isto produz **armas e armaduras de stock** — para equipamento personalizado mais
poderoso, use o Projetor de Armas e o Projetor de Aviões.

**Itens fabricáveis** (construir vários de uma vez):

| Item | Tempo | Materiais |
|------|------|-----------|
| Pistola, Rifle, Canhão Pesado, Canhão Automático | 3–8 dias | Ligas |
| Lançador de Foguetes, Cajado de Atordoamento | 2–8 dias | Ligas + Elério |
| Armadura Pessoal, Fato Leve/Médio/Pesado/Potência/Voador | 6–18 dias | Ligas + Elério |
| Kit Médico | 3 dias | Ligas |

Mais engenheiros = produção mais rápida. Armas de energia (Laser, Plasma) não podem
ser fabricadas — devem ser investigadas, desenhadas e construídas através do
Projetor de Armas, ou recuperadas de aliens.

---

## Equipar Soldados

A partir da aba de Soldados, prima `E` para abrir o ecrã de equipamento.

### Controlos

| Tecla | Ação |
|-----|--------|
| ↑/↓ | Selecionar soldado |
| Tab | Circular itens disponíveis |
| 1 | Slot de arma |
| 2 | Slot de armadura |
| 3 | Slot de inventário |
| Space | Equipar item selecionado |
| G | Abrir projetor de armas |
| A | Auto-equipar todos os soldados |
| Esc | Voltar |

### Slots

- **Slot 1 (Arma):** Arma principal — desenhada ou um rifle/pistola de stock
- **Slot 2 (Armadura):** Armadura corporal — pessoal, fato leve, etc.
- **Slot 3 (Inventário):** Itens extra — granadas, kits médicos, scanners,
  minas de proximidade, amplificadores psi, armas corpo a corpo

### Sobrecarga

Cada item tem peso. O peso total de arma + armadura + inventário é a sua
**sobrecarga**. Maior sobrecarga reduz as suas Unidades de Tempo em batalha
(cerca de 1 UT de penalização por 5 unidades de peso). Mantenha os seus soldados
com carga ligeira para máxima mobilidade.

### Auto-Equipar

Prima `A` para equipar automaticamente cada soldado com a melhor arma e armadura
disponíveis do armazém. O equipamento existente é devolvido ao armazém — uma forma
rápida de reequipar o seu esquadrão após investigar nova tecnologia.

---

## Campo de Batalha

O Campo de Batalha é combate tático por turnos. Você controla um esquadrão de
soldados contra forças alienígenas num mapa de grelha 50×50.

### Estrutura de Turnos

1. **Turno do Jogador** — Mova e atue com cada soldado usando Unidades de Tempo (UT)
2. **Turno Aliens** — Os aliens atuam usando as suas próprias pools de UT
3. Repita até um lado ser eliminado

### Unidades de Tempo (UT)

Cada ação custa UT. As UT restauram totalmente no início de cada turno do jogador.

| Ação | Custo aprox. de UT |
|--------|-----------------|
| Mover (por tile) | 4 |
| Agachar | 4 |
| Disparar arma | Varia por arma (mirado=base, rajada=1,5×, auto=2×) |
| Recarregar | 8 |
| Lançar granada | 20 |
| Usar kit médico | 25 |
| Ataque psi | 20 |

A pool de UT de um soldado começa em 45–55 e pode crescer com a experiência (máx ~80).

### Modos de Disparo

As armas podem ter múltiplos modos de disparo. Prima **Tab** para circular e verifique
o modo mostrado na barra lateral.

| Modo | Custo | Precisão | Disparos | Quando usar |
|------|------|----------|--------|-------------|
| **Mirado** | UT base | Melhor | 1 tiro | Longo alcance, alvos de alto valor |
| **Rajada** | 1,5× UT | -10% | 3 tiros | Médio alcance, supressão |
| **Auto** | 2× UT | -20% | Toda a munição restante | Perto, emergências |

Nem todas as armas suportam todos os modos. Rifles e rifles laser suportam rajada;
apenas poucas armas suportam disparo automático.

### Combate

Fatores de combate:
- **Precisão** depende da perícia do soldado, distância ao alvo, cobertura e modo
  de disparo usado
- **Agachar** dá um bónus de precisão e reduz o dano recebido
- **Cobertura** — paredes bloqueiam 80% do dano, árvores 60%, arbustos 40%,
  cercas 30%. Posicione os seus soldados atrás de cobertura sólida
- **Contorne cobertura** com granadas — explodem numa área e ignoram a redução
  de dano por cobertura

### Linha de Visão

Os soldados só conseguem ver em linha reta. Paredes, árvores e rochas bloqueiam a LdS.
Chão, portas e relva não.

### Objetos e Cobertura

Tiros que passam através de objetos têm o dano reduzido pelo valor de cobertura do
objeto. É aplicada a maior cobertura ao longo da linha de fogo.

| Objeto | Cobertura % |
|--------|---------|
| Parede / Parede OVNI | 80% |
| Rocha | 70% |
| Árvore | 60% |
| Mobília OVNI | 50% |
| Arbusto | 40% |
| Fumo Denso | 40% |
| Cerca | 30% |
| Entulho | 20% |

### Granadas

- Alcance: ~6 tiles
- Dano: baseado na força, com salpicos em área
- Destrói paredes, árvores, rochas e cercas dentro do raio de explosão
- Cria nuvens de fumo que bloqueiam a LdS em alta densidade

### Kit Médico

Cura 10 HP por uso (15 HP com a vantagem de Sano de Campo), custa 25 UT.

### Missões Noturnas

Noite (antes das 6:00 ou depois das 18:00):
- Precisão mais baixa (cerca de 75% do dia)
- Alcance de visão reduzido (de 20 para ~10 tiles)
- Soldados brilham com calor, aliens brilham azul fraco

### Modos de Visão

Prima `V` para circular: **Normal → Visão Noturna → Térmica → Normal**
- Visão Noturna: sobreposição de fósforo verde com estática
- Térmica: entidades vivas brilham quente, terreno é azul frio

### Combate Psi

Requer uma arma Amplificadora Psi e uma instalação Lab Psi. O sucesso depende da
perícia psi do seu soldado vs a força psi do alvo. Um ataque psi bem-sucedido
entraria o alvo em pânico — ele perde o turno.

Soldados numa base com Lab Psi podem ganhar perícia psi ao longo do tempo (até ~80).
Investigação de Controlo Mental concede um grande impulso psi a todos os soldados.

### Modificadores de Missão

Modificadores aleatórios que alteram cada batalha:

| Modificador | O que acontece |
|----------|--------------|
| Operações Noturnas | Batalha noturna forçada, saque extra |
| Reforços | Aliens extra chegam no turno 4 |
| Limite de Tempo | 15 turnos para eliminar todos os aliens |
| Resgate de VIP | Proteger um VIP, dinheiro extra se sobreviver |
| Armadilhada | Mais granadas e minas no mapa |
| Nevoeiro Espesso | Alcance de visão reduzido em 40% |
| Emboscada Alien | Aliens começam em posições de vigilância |
| Baixa Visibilidade | Precisão reduzida para todas as unidades |
| Terreno Elevado | Posições elevadas dão bónus de precisão |

### Tempo

O tempo afeta o combate com base na localização da missão:

| Tempo | Efeito |
|---------|--------|
| Chuva | Precisão mais baixa, fogo espalha mais devagar |
| Vento | Fogo espalha mais rápido, granadas podem derivar |
| Neve | Movimento custa mais em neve profunda |
| Nevoeiro | Precisão mais baixa, visão reduzida |
| Tempestade | Chuva + vento combinados |
| Frio | Pequena penalização de precisão |

### Relatório Pós-Ação

Após cada batalha, vê:
- Resultado (Vitória / Derrota)
- Aliens mortos e soldados perdidos
- Saque recuperado e prisioneiros capturados
- Fundos ganhos
- Ganhos de estatística por soldado ou marcador "KIA"

Prima **Enter**, **Space** ou **Esc** para descartar.

---

## Armas e Equipamento

### Projetor de Armas Personalizado

Prima `G` a partir da Base, Soldados ou ecrã de Equipamento para abrir o **Projetor
de Armas**. Esta é a forma principal de criar armas para os seus soldados. Escolha
um modelo base e personalize cada componente:

| Componente | Opções | O que afeta |
|-----------|---------|-----------------|
| **Base** | Pistola / Rifle | Dano inicial, alcance, precisão, custo de UT |
| **Cano** | Curto / Padrão / Longo / Estendido | Alcance, precisão, custo de UT, peso |
| **Ótica** | Nenhuma / Miras de Ferro / Mira Telescópica / Ótica Avançada | Precisão, custo de UT, peso |
| **Modo de Tiro** | Semi-Auto / Full-Auto | Modo automático (dispara mais rápido, menos preciso) |
| **Munição** | Padrão / Perfurante / Incendiária / Explosiva | Modificador de dano, custo de UT, peso |
| **Coronha** | Nenhuma / Leve / Pesada | Precisão, custo de UT, peso |

Cada componente afeta o dano, precisão, custo de UT, alcance e peso da arma.
O painel de pré-visualização mostra a arma montada como arte ASCII colorida e
apresenta as suas estatísticas finais. Os desenhos são guardados como itens
personalizados disponíveis no ecrã de Equipamento.

**Dica:** Comece com uma base de Rifle para a maioria dos propósitos. Canos longos
e miras telescópicas melhoram a precisão à distância. Munição explosiva atinge bem
mas custa UT extra.

### Armas de Stock

Estes itens base estão disponíveis desde o início e podem ser fabricados:

| Tipo | Dano | Munição | Notas |
|------|--------|------|-------|
| Pistola | Ligeiro | Balística | Precisa de recarga, peso baixo |
| Rifle | Médio | Balística | Emissão padrão, suporta rajada |
| Canhão Pesado | Alto | Balística | Lento, pesado, atinge forte |
| Canhão Automático | Médio | Balística | Opção full-auto |
| Lançador de Foguetes | Muito Alto | Explosiva | Dano em área |

### Armas de Energia

Investigadas mais tarde — nunca precisam de recarga:

| Tipo | Dano | Notas |
|------|--------|-------|
| Pistola Laser | Ligeiro | Arma de energia inicial |
| Rifle Laser | Médio | Suporta rajada, nunca recarrega |
| Pistola de Plasma | Médio | Arma alienígena, nunca recarrega |
| Rifle de Plasma | Alto | Arma alienígena, nunca recarrega |
| Plasma Pesado | Muito Alto | Arma alienígena de topo |

### Munição e Recarga

- **Armas balísticas** precisam de recarga — prima `R` em combate
- **Armas de energia** (Laser, Plasma) nunca precisam de recarga
- **Consumíveis** (granadas, kits médicos) são usados do inventário

### Modos de Disparo

Veja [Campo de Batalha → Modos de Disparo](#modos-de-disparo) para detalhes.

### Itens de Inventário

Os soldados podem transportar itens extra no seu slot de inventário:
- **Granadas** — explosivo arremessado com dano em área
- **Kits Médicos** — cure-se ou a um aliado adjacente
- **Scanners de Movimento** — detetam inimigos próximos
- **Minas de Proximidade** — colocadas no chão, detonam quando um inimigo passa por cima
- **Amplificadores Psi** — permitem ataques psi (requer perícia psi)
- **Armas corpo a corpo** — Cajado de Atordoamento para neutralizações não letais

Cada item de inventário adiciona peso e aumenta a sobrecarga, reduzindo as suas
UT disponíveis em batalha. Empacote com inteligência.

### Itens Procedurais

Cada partida gera armas e armaduras únicas com base na espécie alienígena encontrada.
Estes têm nomes e estatísticas aleatórias — cada jogo é diferente.

**Armas procedurais:** 2–3 armas com tipos de dano correspondentes à espécie alienígena.
**Armaduras procedurais:** 1–2 peças de armadura com proteção correspondente aos tipos
de dano alienígenas.

Estes itens são adicionados automaticamente ao seu armazém no início do jogo.

---

## Armadura

| Armadura | Defesa | Penalização de UT | Notas |
|--------|---------|------------|-------|
| Nenhuma | 0 | Nenhuma | Padrão |
| Armadura Pessoal | 10 | Nenhuma | Padrão do início do jogo |
| Fato Leve | 20 | -5% UT | Boa opção a meio do jogo |
| Fato Médio | 30 | -10% UT | Proteção forte |
| Fato Pesado | 40 | -15% UT | Defesa máxima, penalização pesada |
| Fato de Potência | 50 | -10% UT | Armadura de fim de jogo |
| Fato Voador | 45 | -5% UT | Quase fim de jogo, mais leve que o de Potência |

Maior defesa reduz o dano recebido, mas fatos mais pesados custam Unidades de Tempo.

---

## Aliens

### Espécies Procedurais

Cada jogo gera 5–7 espécies alienígenas únicas a partir de uma semente. Cada espécie
tem 2–5 variantes de posto (Soldado → Navegador → Comandante → Elite → Soberano).

As espécies diferem em:
- **Tipo de dano** — o tipo de dano que causam
- **Resistências e fraquezas** — alguns são fracos a plasma, outros a explosivos
- **Preferência de arma** — postos baixos usam pistolas, postos altos usam armas pesadas
- **Morfologia** — plano corporal físico que afeta estatísticas e resistências

Isto significa que **cada partida apresenta ameaças alienígenas diferentes**. Uma
partida pode ter uma espécie psionicamente pesada, outra pode ter predadores corpo a
corpo fracos a explosivos.

### Morfologia

A morfologia determina a forma física de um alien. Fatores-chave:

**Membros:**
- Braços (0–6): Menos braços = pior precisão, mais braços = melhor estabilidade ou uso dual
- Pernas (0–8): Mais pernas = mais rápido mas alvo maior; zero pernas = flutua, mais difícil de acertar

**Tipos de corpo e as suas resistências:**
- **Carne de Carbono:** +Resistência Cinética, -Fraqueza Explosiva
- **À Base de Silício:** +Laser/+Resistência Plasma, -Fraqueza Explosiva, refletivo
- **Gasoso:** Imune a cinético, fraco a plasma, pode atravessar paredes
- **Cristalino:** Boa resistência geral, muito fraco a explosivos, estilhaça ao morrer
- **Amorfo:** +Resistência Psi, regenera HP a cada turno
- **Mecânico:** Imune a psi, +Resistência Plasma, -Fraqueza Laser, autodestrói-se
- **Bio-Sintético:** Resistências equilibradas, cura aliens adjacentes
- **Nanotecnológico:** +Resistência Cinética, pode reviver ao morrer

**Sentidos:**
- **Visão:** Afeta precisão — multi-espectro ignora fumo/escuridão
- **Audição:** Ecolocalização deteta unidades através do fumo a curta distância
- **Sentido Térmico:** Deteta unidades vivas independentemente da cobertura a curta distância
- **Sentido Psiónico:** Aumenta psi, deteta humanos controlados mentalmente
- **Sentido Químico:** Bónus de precisão contra alvos feridos

### Níveis de Conhecimento

À medida que encontra aliens, a inteligência melhora:

| Nível | O que aprende |
|-------|----------------|
| Desconhecido | Nome aparece como "???" |
| Avistado | Nome e ícone revelados |
| Morto | Estatísticas e resistências reveladas |
| Autopsiado | Lore completo e fraquezas detalhadas |

### IA Alien

Os aliens patrulham até avistarem um humano, depois atacam. Comportamentos incluem:
- **Procurar** — mover em direção à última posição conhecida por alguns turnos
- **Fugir** — afastar-se quando gravemente ferido e com pouca coragem
- **Adaptar** — os aliens estudam as suas táticas ao longo das missões.
  Atira de longe? Eles avançam sobre si. Usa granadas? Eles espalham-se.
  Flanqueia frequentemente? Eles postam supressores.

### Escalada de Equipamento

Os aliens obtêm melhor equipamento à medida que a campanha progride:
- **Meses iniciais:** Pistolas de plasma, armadura básica
- **Meio da campanha:** Rifles de plasma, plasma pesado, canhões alienígenas
- **Fim da campanha:** Armas e armaduras alienígenas de topo

### Captura Alien

Use um **Cajado de Atordoamento** (corpo a corpo, $2K para fabricar) para deixar aliens
inconscientes. Se o dano de atordoamento exceder o HP deles, caem inconscientes e podem
ser recolhidos após a missão — desde que tenha Contenção de Aliens com capacidade livre.

Aliens capturados podem ser interrogados a partir do ecrã de Investigação (tecla `I`):
- A interrogação pode concluir uma autópsia ativa instantaneamente
- Ou conceder um bónus de progresso à investigação atual
- Requer pelo menos um Laboratório

---

## Postos e Progressão dos Soldados

### Postos

Os postos desbloqueiam à medida que o seu efetivo total cresce:

| Posto | Desbloqueia quando o efetivo atinge |
|------|----------------------------|
| Recruta | Sempre disponível |
| Soldado | Sempre disponível |
| Cabo | 4 soldados |
| Sargento | 8 soldados |
| Tenente | 14 soldados |
| Capitão | 22 soldados |
| Major | 30 soldados |
| Coronel | 40 soldados |

### Crescimento de Estatísticas

Os soldados melhoram através de **experiência por ação** durante a batalha:
- **Disparo** → melhora Precisão
- **Reações** → melhora Reações
- **Corpo a corpo** → melhora Força
- **Coragem** → melhora Coragem (por resistir ao pânico)
- **Perícia psi** → melhora Perícia Psi e Força Psi

Após cada missão, o XP acumulado é convertido em ganhos de estatística. Soldados que
ganharam XP também obtêm crescimento geral de "halo" em direção aos limites de HP, UT
e Força. Os limites são aproximadamente: UT 80, HP 60, Precisão 120, Reações 100,
Coragem 100, Força 70, Psi 100.

### Fadiga e Feridas

- **Soldados feridos** não podem ser destacados até curar (2 HP/dia de recuperação)
- **Fadiga:** Batalhas causam 1–5 dias de fadiga
- Instalações de cura e Quartos de Habitação aceleram a recuperação

### Feridas Fatais

Em batalha, tiros podem causar feridas fatais e hemorragia. A hemorragia drena HP a
cada turno — ponha um kit médico neles rápido. Feridas sobreviventes tornam-se dias
de recuperação após a missão.

### Moral

Os soldados recuperam moral a cada turno. Baixa moral pode desencadear pânico (salta turno).
Resistir ao pânico constrói XP de coragem.

### Vantagens

Cada subida de posto concede uma vantagem aleatória:

| Vantagem | Efeito |
|------|--------|
| Reflexos Relâmpago | +10 Reações |
| Atirador de Elite | +Precisão a longo alcance |
| Granadeiro | Maior salpico de granada |
| Sano de Campo | Kit médico cura mais |
| Ferro Will | +Perícia Psi e +Força Psi |
| Pontaria Firme | +Precisão quando estático |
| Especialista Corpo a Corpo | +Precisão a curto alcance |
| Perito em Vigilância | +Precisão de fogo de reação |
| Demolições | +Dano de granada |
| Catador | +Saque das batalhas |
| Resistente | +5 HP máx |
| Aprendiz Rápido | +Ganho de XP |

### Memorial

Soldados mortos em ação são registados no Memorial do jogo.
Pode vê-lo para honrar os caídos.

---

## Guardar/Carregar

| Tecla | Ação |
|-----|--------|
| F5 | Abrir selecionador de slot de gravação |
| F9 | Abrir selecionador de slot de carregamento |

As gravações incluem: tempo de jogo, fundos, estado de pausa, atividade alienígena,
estado da base, OVNIs, missões ativas, semente de espécies procedurais e níveis de
conhecimento alienígena. A semente garante que as mesmas espécies alienígenas
regeneram ao recarregar.

**Gravação automática:** Se ativada nas Opções, o jogo grava automaticamente periodicamente.

---

## Referência de Teclas

### Geomapa

| Tecla | Ação |
|-----|--------|
| Setas | Mover câmara |
| j/k | Navegar lista de regiões |
| Space | Pausar/retomar |
| 1–4 | Velocidade do tempo |
| B | Abrir base |
| L | Lançar interceptor |
| A | Autoresolver OVNI mais próximo |
| M | Responder a missão |
| R | Enviar transporte para local de acidente |
| C | Circular para próxima base |
| N | Construir nova base ($500K) |
| T | Abrir ecrã de transferência |
| E | Abrir enciclopédia |
| V | Alternar sobreposição de radar |
| F5 | Guardar |
| F9 | Carregar |
| Q | Sair |
| ? | Ajuda |

### Gestão da Base

| Tecla | Ação |
|-----|--------|
| 1–6 | Trocar abas |
| j/k | Navegar itens |
| B | Construir instalação |
| S | Vender instalação |
| H | Contratar soldado |
| E | Abrir ecrã de equipamento |
| G | Abrir projetor de armas |
| D | Dispensar soldado / Projetor de Aviões (Hangares) |
| Esc | Voltar ao geomapa |

### Ecrã de Equipamento

| Tecla | Ação |
|-----|--------|
| ↑/↓ | Selecionar soldado |
| Tab | Circular itens disponíveis |
| 1 | Slot de arma |
| 2 | Slot de armadura |
| 3 | Slot de inventário |
| Space | Equipar item selecionado |
| G | Abrir projetor de armas |
| A | Auto-equipar todos os soldados |
| Esc | Voltar |

### Projetor de Armas

Prima `G` a partir da Base, Soldados ou ecrã de Equipamento.

| Parâmetro | Opções | Efeito |
|-----------|---------|--------|
| Cano | Curto / Padrão / Longo / Estendido | Alcance, precisão, peso, custo de UT |
| Ótica | Nenhuma / Miras de Ferro / Mira Telescópica / Avançada | Precisão, peso, custo de UT |
| Modo de Tiro | Semi-Auto / Full-Auto | Modo automático |
| Munição | Padrão / Perfurante / Incendiária / Explosiva | Dano, peso, custo de UT |
| Coronha | Nenhuma / Leve / Pesada | Precisão, peso, custo de UT |

### Projetor de Aviões (Interceptores Personalizados)

Todos os interceptores são desenhados e construídos através do **Projetor de Aviões**.
Prima `D` a partir da aba de Hangares para abri-lo. Configure a sua aeronave:

| Parâmetro | Intervalo | O que afeta |
|-----------|-------|-----------------|
| **Comprimento** | Curto (3) → Longo (7) | Pontos de casco, massa, velocidade |
| **Envergadura** | Curto (1) → Largo (4) | Manobrabilidade, massa |
| **Motores** | 1–3 | Velocidade, capacidade de combustível, massa |
| **Combustível** | 20–100 | Alcance operacional |
| **Arma** | Canhão / Stingray / Avalanche / Plasma | Poder de fogo, peso, custo |
| **Blindagem** | Nenhuma / Liga Leve / Liga Pesada / Revestimento Alien | Bónus de casco, redução de dano, massa |

O projetor calcula estatísticas derivadas (velocidade, poder de fogo, casco,
rácio massa/empuxo) a partir da sua configuração e mostra uma pré-visualização ASCII
colorida. Projetos mais pesados são mais resistentes mas mais lentos — equilibre
durabilidade contra velocidade de interceção.

**Armas de avião:**

| Arma | Dano | Precisão | Alcance | Cadência | Custo |
|--------|--------|----------|-------|-----------|------|
| Canhão | 15 | 85% | 25 | 3 tiros | $5K |
| Stingray | 25 | 70% | 45 | 2 tiros | $8K |
| Avalanche | 40 | 55% | 60 | 1 tiro | $12K |
| Plasma | 60 | 50% | 50 | 1 tiro | $20K |

**Blindagem de avião:**

| Blindagem | Bónus de Casco | Redução de Dano | Custo |
|--------|------------|------------------|------|
| Nenhuma | 0 | 0% | Grátis |
| Liga Leve | +10 | 10% | $8K |
| Liga Pesada | +25 | 25% | $18K |
| Revestimento Alien | +40 | 40% | $35K |

### Campo de Batalha

| Tecla | Ação |
|-----|--------|
| Setas / WASD / hjkl | Mover cursor |
| Space / Enter | Selecionar unidade / confirmar |
| q | Circular soldados |
| f | Disparar arma |
| Tab | Circular modo de disparo |
| r | Recarregar |
| e / n | Terminar turno |
| c | Agachar |
| g | Lançar granada |
| m | Modo de movimento |
| h | Usar kit médico |
| p | Ataque psi |
| y | Scanner de movimento |
| t | Colocar mina de proximidade |
| v | Circular modo de visão |
| o | Opções |
| ? | Ajuda |
| Esc | Cancelar / desselecionar |

### Controlos Táteis Móveis

No navegador com ecrã estreito (cols < 100) ou quando `touch_mode` está ativado:

| Gesto | Ação |
|---------|--------|
| Toque | Selecionar, mover, disparar |
| Pressão longa (500ms) | Cancelar |
| Arrastar vertical | Scroll |

Um botão `[=]` abre um menu de controlo tátil no ecrã.

---

## Dicas e Estratégia

### Início do Jogo

1. Investigue **Ligas Aliens** primeiro — desbloqueia Armas e Armadura Laser.
2. Construa um segundo **Radar** — mais deteção, mais financiamento mensal.
3. Contrate 2–4 soldados extra para preencher os seus esquadrões.
4. Use o **Projetor de Armas** para criar rifles personalizados — pode construir
   melhores armas que os modelos de stock com os componentes certos.
5. Não ignore autópsias — algumas tecnologias de armas exigem-nas.
6. Desenhe os seus interceptores no **Projetor de Aviões** — um projeto
   equilibrado (comprimento/mediana médio, 2 motores, mísseis Stingray) supera
   os interceptores padrão.

### Combate

- **Use cobertura** — paredes (80%) > rochas (70%) > árvores (60%) > arbustos (40%)
- **Agache** antes de disparar para melhor precisão e redução de dano
- **Granadas** contornam cobertura e destroem paredes — perfeitas para inimigos entrincheirados
- **Aprenda resistências alienígenas** — consulte a enciclopédia após as primeiras mortes
- **Não se estenda demasiado** — aliens obtêm disparos de reação quando se move em sua LdS
- **Mantenha um médico** — um soldado com kit médico pode salvar vidas
- **Gira sobrecarga** — não sobrecarregue soldados com equipamento pesado

### Economia

- Venda cadáveres e saque alienígena excedentes por dinheiro
- Os salários mensais acumulam — equilibre o seu efetivo contra a receita
- Missões de Conselho pagam bónus de $100K — priorize-as
- Fabrique itens para vender com lucro no início do jogo

### Caminho de Investigação

Ligas → Armas Laser → Armadura Pessoal → Autópsias → Elério → Armas de Plasma

Meio do jogo: Fato Médio, Plasma Pesado.
Fim do jogo: Fato de Potência/Voador, Controlo Mental.

### Construção de Base

- Radares pagam-se a si próprios (+$50K/mês cada)
- Construa Armazém cedo — enche rápido
- Contenção de Aliens é necessária para capturas vivas e bónus de interrogação
- Instalações adjacentes reforçam-se mutuamente — planeie o layout da base
- Construa um Lab Psi se quiser capacidades psi


