# 7. Documentation layout: README, CURSOR.md, and `docs/`

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The project needs a user README ([Make a README](https://www.makeareadme.com/)), agent/implementation rules, architecture decisions, and use cases for ported modules. Dumping everything into `CURSOR.md` or the README makes both unusable.

## Considered options

1. **Only README** — not enough for agents or ADRs.
2. **README + root ARCHITECTURE.md + CLAUDE/CURSOR.md** — splits docs across the repo root.
3. **README (users) + CURSOR.md (agents) + `docs/architecture` (ADRs) + `docs/use-cases` (after each module is ported) + `docs/wiki` (operator how-tos).**

## Decision

Use option 3. All long-form product/architecture docs besides README and CURSOR.md live under `docs/`.

**README** stays the short entry (what the app is, install, OBS URLs, status). Step-by-step operator guides live in **`docs/wiki/`** and the GitHub wiki. Do not duplicate wiki how-tos (YAML examples, curl, quota, config paths) in the README.

Inside wiki pages:

* Link other wiki pages **without** `.md`: `[Configuration](Configuration)`, `[Getting Started](Getting-Started#which-messages-you-will-see)`.
* Link screenshots with raw URLs on `main` (`https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/home.png`). Relative `../screenshots/` paths do not resolve on the wiki.
* `_Sidebar.md` and `Home.md` are GitHub wiki special pages (sidebar + landing).

Write a use-case file **only after** that module exists in this repo. Until then, list the scenario in `docs/use-cases/INDEX.md` as Planned. When a module is ported, add `uc-NN-….md` and set the index status to Ported.

## Consequences

### Positive

* Users start at README; operator how-tos are in `docs/wiki/`; agents start at CURSOR.md; decisions are numbered ADRs.
* Use cases stay honest (no fiction about unported code).

### Negative and risks

* Four places to update when a user-visible first-run contract changes (README + `docs/wiki/` + CURSOR.md + ADR if the decision changed).

### Neutral

* Use-case prose follows the template in `docs/use-cases/_TEMPLATE.md` (actors, goal, preconditions, happy path, alternatives, postconditions).
* Operator how-tos in `docs/wiki/` use GitHub wiki file names (`Getting-Started.md`). Page links omit `.md`. Screenshots use `raw.githubusercontent.com/.../main/docs/screenshots/`.
* README window showcase is `window.gif`, generated locally with `make screenshots` via Playwright + gifenc and displayed from `docs/screenshots/window.gif`.
