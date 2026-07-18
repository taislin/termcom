# Início Rápido do termcom

Um demake X-COM em ASCII para o seu terminal. Comande a defesa da humanidade contra a invasão alienígena.

## Executar

```bash
go run ./cmd/termcom      # ou: make run
```

## Ciclo de Jogo

1. **Geomapa** -- OVNIs voam em direção às cidades. Detete e intercepte-os.
2. **Intercetar** -- Lance caças (L) ou autoresolva (A) para abater OVNIs.
3. **Batalha** -- Destaque para locais de acidente (R). Entre em combate tático.
4. **Base** -- Investigue tecnologia alienígena, fabrique equipamento, contrate/equipe soldados.
5. **Repetir** -- Vença 10 batalhas, depois assalte Cydonia para salvar a Terra.

Perde se a Atividade Aliens atingir 100%.

## Teclas Essenciais (Geomapa)

| Tecla | Ação |
|-----|--------|
| Space | Pausa |
| 1-4 | Velocidade do tempo |
| L | Lançar interceptor |
| A | Autoresolver OVNI |
| M | Responder a missão |
| R | Enviar transporte para acidente |
| B | Abrir base |
| F5/F9 | Guardar/Carregar |
| Q | Sair |

## Teclas Essenciais (Campo de Batalha)

| Tecla | Ação |
|-----|--------|
| Seta/WASD | Mover cursor |
| Space/Enter | Selecionar/Confirmar |
| F | Disparar arma |
| R | Recarregar |
| Q | Circular soldado |
| E | Terminar turno |
| C | Agachar |
| Esc | Cancelar |

## Estratégia Rápida

- **Início:** Contrate soldados, investigue Ligas Aliens, construa Lab + Oficina
- **Meio:** Armas laser personalizadas (Projetor de Armas) → Armadura Pessoal, expanda bases
- **Fim:** Armas de plasma personalizadas, Fatos de Potência/Voador, treino psi
- Equipe sempre os soldados antes da batalha. Feridos curam 2 HP/dia.
- Desenhe interceptores personalizados no Projetor de Aviões — mísseis Stingray + Blindagem Liga Leve é um bom início.
- Venda artefactos alienígenas excedentes por dinheiro. Instalações de radar aumentam o financiamento.

## Vitória

Vença 10 batalhas terrestres para desbloquear a missão final de Cydonia. Destrua Cydonia para vencer.

Para o manual completo consulte [manual.pt.md](manual.pt.md).
