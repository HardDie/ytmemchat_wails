# 7. Documentation layout: README, CURSOR.md, and `docs/`

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The project needs several kinds of docs.

1. A user README ([Make a README](https://www.makeareadme.com/)).
2. Agent/implementation rules.
3. Architecture decisions.
4. Use cases for ported modules.
5. Dumping everything into `CURSOR.md` or the README makes both unusable.

## Considered options

1. **Only README** — not enough for agents or ADRs.
2. **README + root ARCHITECTURE.md + CLAUDE/CURSOR.md** — splits docs across the repo root.
3. Split docs:
   1. README (users)
   2. CURSOR.md (agents)
   3. `docs/architecture` (ADRs)
   4. `docs/use-cases` (after each module is ported)
   5. `docs/wiki` (operator how-tos)

## Decision

Use option 3.

1. All long-form product/architecture docs besides README and CURSOR.md live under `docs/`.
2. **README** stays the short entry (what the app is, install, OBS URLs, status).
3. Step-by-step operator guides live in **`docs/wiki/`** and the GitHub wiki.
4. Do not duplicate wiki how-tos (YAML examples, curl, quota, config paths) in the README.

Inside wiki pages:

1. Link other wiki pages **without** `.md`.
   1. `[Configuration](Configuration)`
   2. `[Getting Started](Getting-Started#which-messages-you-will-see)`
2. Link screenshots with raw URLs on `main`.
3. Example: `https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/home.png`.
4. Relative `../screenshots/` paths do not resolve on the wiki.
5. `_Sidebar.md` and `Home.md` are GitHub wiki special pages (sidebar + landing).

Use cases:

1. Write a use-case file **only after** that module exists in this repo.
2. Until then, list the scenario in `docs/use-cases/INDEX.md` as Planned.
3. When a module is ported, add `uc-NN-….md` and set the index status to Ported.

## Consequences

### Positive

* Users start at README.
* Operator how-tos are in `docs/wiki/`.
* Agents start at CURSOR.md.
* Decisions are numbered ADRs.
* Use cases stay honest (no fiction about unported code).

### Negative and risks

* Four places to update when a user-visible first-run contract changes (README + `docs/wiki/` + CURSOR.md + ADR if the decision changed).

### Neutral

1. Use-case prose follows the template in `docs/use-cases/_TEMPLATE.md`.
   1. Actors, goal, preconditions, happy path, alternatives, postconditions.
2. Operator how-tos in `docs/wiki/` use GitHub wiki file names (`Getting-Started.md`).
3. Page links omit `.md`.
4. Screenshots use `raw.githubusercontent.com/.../main/docs/screenshots/`.
5. README window showcase is `window.gif`.
6. Generate it locally with `make screenshots` via Playwright + gifenc.
7. Display it from `docs/screenshots/window.gif`.
