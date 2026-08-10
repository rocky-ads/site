# Rocks at Rocky Ads

## Community Moderation Without a Moderation Desk

**Document type:** Product design, threat, and legal analysis  
**Audience:** Product, engineering, trust & safety, and legal reviewers  
**Scope:** How Rocky Ads rocks work end-to-end; the philosophy and motivation behind scarcity-based, peer-visible disputes; known flaws; realistic abuse and exploit paths; and legal tensions that follow from self-policing with public rock signals and AI dispute assessments.  
**Based on:** Rocky Ads site implementation (Go / Fiber / Postgres / Grok rock opinions / encrypted conversation journals), August 2026.  
**Companions:** [DOC_PHONE_ONLY_IDENTITY_AND_SECURITY.md](DOC_PHONE_ONLY_IDENTITY_AND_SECURITY.md) (identity / SMS); [DOC_LOCAL_LAUNCH_CAR_TRUCK_PARTS.md](DOC_LOCAL_LAUNCH_CAR_TRUCK_PARTS.md) (go-to-market: rocks as trust signal, not a substitute for liquidity).

---

## Abstract

Rocky Ads treats classifieds trust as a **scarce, peer-enforced signal**, not as a staffed queue of tickets. Every member receives a fixed allotment of **rocks**. Throwing a rock on a conversation marks a dispute, makes that thread publicly addressable to other logged-in members, increments a visible rock count on either the **ad** (buyer/inquirer threw) or the **counterparty’s profile** (seller/owner threw), and—when enough ad-bound rocks accumulate—**removes the listing from search**.

The design is deliberately anti-bureaucratic. Terms and FAQ push listing problems into the product: throw a rock, talk it out, unthrow when resolved. An optional **Grok-generated rock opinion** paraphrases the dispute for bystanders so they need not read the raw transcript as the primary UI. Ads are **assumed good until rocked**.

That bet buys speed, honesty about “who cares enough to spend a rock,” and a marketplace metaphor that matches the brand. It also creates sharp edges: rock scarcity can be gamed or exhausted; public disputes collide with privacy expectations about private messages; FAQ/Terms copy about “the seller can return the rock” does not match code (only the thrower can unthrow); AI opinions leave the trust boundary and can be wrong; and self-policing is an incomplete answer to statutory takedown, CSAM, and other non-negotiable removal duties. This paper maps the mechanism, the philosophy, the failure modes, and the legal posture without pretending rocks are a complete trust system.

---

## 1. Motivation: why rocks exist

Classic newspaper classifieds had almost no platform moderation: the paper printed the ad; buyers and sellers sorted risk in the real world. Modern marketplaces inverted that. Facebook Marketplace, Craigslist flags, and app-store marketplaces accumulate report queues, automated classifiers, and opaque enforcement. Users learn that “Report” is slow, inconsistent, or captured by whoever can spam the button.

Rocky Ads starts from a different product thesis (also reflected in the local-launch notes):

| Pain on FB / CL | Rocky Ads angle |
|-----------------|-----------------|
| No shared trust signal | Rocks mark disputed ads / actors; ads assumed good until rocked |
| Fake / throwaway reporters | Phone-gated accounts raise the cost of sockpuppets; rocks themselves are scarce |
| Hidden numbers / deposit scams | Contact stays in-app; rocks attach to the conversation, not a public phone |
| Flaky chat | Dedicated messaging + rock as a first-class dispute object |

Phone verification and rocks are **complements**, not substitutes. Phone cost slows mass spam; rocks make *individual* complaints expensive for the complainant and consequential for the listing. Neither invents liquidity. Rocks are the retention and trust story once inventory exists—not a cold-start engine.

The brand metaphor is intentional: a rock is heavier than a “like” or a flag emoji. You only get a few. Spending one should feel like staking reputation on a claim.

---

## 2. Philosophy

Several design commitments sit behind the mechanism:

1. **Innocent until rocked.** Search and browse do not require prior “verified seller” theater. A listing with zero rocks is treated as fine for discovery. Distrust is earned by peers, not pre-assigned by the platform.

2. **Scarcity over infinite flags.** `MaxOutstandingRocks = 3` at signup. A user may have at most three rocks in flight. That forces triage: you cannot carpet-bomb every ad you dislike.

3. **Dispute is a conversation, not a ticket.** Throwing a rock is bound to a buyer–seller thread. The FAQ frames resolution as working it out with the other party. The platform’s job is to make the dispute *visible and costly*, not to sit as judge of first resort.

4. **Binding depends on who throws.**  
   - **Inquirer throws** → rock is **ad-bound** (`rock_thrower_id = inquirer_id`). Visible as rock icons on the ad; counts toward `ads.rock_count` and search exclusion.  
   - **Owner throws** → rock is **user-bound** (`rock_thrower_id = owner_id`). Visible on the inquirer’s profile as rocks “against” that user. Does **not** bump the ad’s rock count.

5. **Publicity is the enforcement.** A rock is not a private DM to staff. It opens the conversation to other authenticated members for viewing (and, via rock icons, surfaces an AI assessment). Shame and search demotion do the work that a moderator queue would otherwise do.

6. **AI as optional arbitrator, not court.** Rock opinions paraphrase and score fault provisionally. Copy and prompts insist this is community guidance, not legal advice. Redaction tries to keep phones, emails, street addresses, and participant names out of the published summary.

7. **Operator reserve powers remain.** Terms still allow Rocky Ads to remove content or restrict accounts when needed for law or service protection. Rocks are the *default* reporting path (“do not email us to report listing problems”), not a waiver of emergency removal.

---

## 3. How rocks work

### 3.1 Lifecycle

```text
Signup → user receives MaxOutstandingRocks (3) conceptual allotment
                ↓
Participant opens / creates conversation on an active ad
                ↓
Throw rock button → confirm modal (GET)
                ↓
Confirm: remaining rock icons, caution (incl. unthrow),
         canned reason radios, optional assessment preview
                ↓
Throw rock (POST + reason) → rock_thrower_id + rock_thrown_at set
                  → journal appends "rock thrown" with reason body
                  → store rock_opinions (async; includes images)
                  → if inquirer: SyncAdRockCount
                ↓
UI: rock icons on ad (ad-bound) or user (user-bound)
    Bystanders clicking icons → rock opinion modal
    Participants → full conversation modal
                ↓
Unthrow (DELETE, thrower only) → clear rock fields
                              → journal "rock unthrown"
                              → invalidate rock_opinions row
                              → resync ad rock_count if ad-bound
                              → thrower regains a rock in allotment
```

Constraints enforced in code:

| Rule | Behavior |
|------|----------|
| Participants only | Non-participants cannot throw |
| One rock per conversation | `ErrRockAlreadyThrown` if already set |
| Outstanding cap | `GetUserRockCount` ≥ 3 → `ErrMaxRocksReached` |
| Reason required | Canned `policy` / `conduct` / `deal` only |
| Messaging open | Inactive/deleted ads or deleted counterparties block throw |
| No self-rock on own ad | `ThrowRockOnAdHandler` forbids rocking your own listing |
| Unthrow authority | Only `rock_thrower_id` may clear the rock |

Force-unthrow paths exist for lifecycle hygiene: deleting an ad, deleting a user, or related cleanup clears outstanding rocks and appends journal unthrow events so counts and public state do not dangle.

### 3.2 Data model

On `conversations`:

- `rock_thrower_id` NULL = private thread; NOT NULL = public rock state.  
- Schema comment encodes binding: owner throw → bound to inquirer; inquirer throw → bound to ad.  
- `rock_thrown_at` timestamps the active rock.  
- Encrypted `journal` records throw/unthrow as first-class events alongside messages. Thrown rocks store a canned reason code in the journal body (`policy`, `conduct`, or `deal`).

On `ads`:

- `rock_count` is a denormalized count of **ad-bound** rocks only (inquirer = thrower). Owner-thrown rocks do not inflate listing rock counts.

On `rock_opinions`:

- Cached LLM output keyed by `conversation_id`: summary, assessment 1–10, assessment detail, resolution, reasoning.  
- Invalidated on unthrow (and when ad-level invalidation runs).

### 3.3 Search and discovery

Vector / metadata search adds `rock_count <= MaxRockCount` where `MaxRockCount = 2`. So an ad remains discoverable with 0–2 ad-bound rocks and drops out of search listings once **more than two** rocks are outstanding against it (i.e., at three). Icons still communicate trouble before exclusion; exclusion is the hard consequence.

Embeddings include natural-language rock context (“no reported disputes” vs “N reported disputes”) so similarity search is not blind to dispute density.

### 3.4 UX surfaces

- **Throw** opens a confirm modal (remaining rock icons, caution including unthrow, required canned reason, optional dispute assessment preview) before POST.  
- **Unthrow** remains a direct control for the thrower in the conversation modal.  
- **Rock icons** on ads and user profiles; each icon maps to an ordinal public rock conversation.  
- **Rock opinion modal** for bystanders (and for participants via a dedicated opinion route).  
- **Remaining-rocks badge** in chrome as a scarcity reminder.  
- FAQ `/faq/rocks` and Terms “Reporting and removal” / Privacy “Content reports” all point users at rocks as the in-product report path.

### 3.5 Rock opinions (Grok)

Before throw, **Preview** generates an ephemeral opinion (not stored) with the selected reason. After throw, `GenerateAndStore` / `GetOrGenerate` persist an opinion:

1. Loads ad + tags + formal facets + journal messages (+ selected reason).  
2. Redacts likely phones, emails, street patterns, and participant usernames from message text before the model sees them.  
3. Attaches up to four listing images (`480w`) as base64 multimodal parts when the image store is configured.  
4. Calls Grok with a system prompt framing a neutral Owner/Inquirer arbitrator (text + images).  
5. Parses JSON; refuses to store if sanitized fields still look like they contain PII patterns.  
6. Stores the opinion for subsequent viewers (preview path skips store).

Invalidation runs on unthrow, ad edit, and ad image confirm.

Assessment scale: **1** = inquirer clearly in the right … **5** = balanced … **10** = owner clearly in the right. Resolution text suggests concrete next steps. The product presents this as provisional community guidance.

---

## 4. What rocks are not

Clarity matters for both product honesty and legal framing:

- **Not KYC or identity proof.** A rock does not verify who is “right”; phone OTP did the human-cost check at signup.  
- **Not a staffed trust & safety queue.** Default path is peer dispute; operator intervention is reserved.  
- **Not a court, insurance, or escrow.** Rocky Ads does not hold funds or adjudicate property title.  
- **Not a complete CSAM / illegal-goods pipeline.** Those categories require operator/legal removal regardless of rock count.  
- **Not a substitute for liquidity.** Launch docs warn explicitly against expecting rocks + phone-only to invent supply.

---

## 5. Flaws and design tensions

These are inherent or currently observable tensions in the running system—not a claim that the feature “doesn’t work,” but that it works *as a particular kind of social technology* with predictable failure modes.

### 5.1 Scarcity cuts both ways

Three outstanding rocks protect against flag spam. They also mean a good-faith user who encounters four bad listings must choose which three to rock—or wait until earlier disputes unthrow. Coordinated scammers can burn a victim’s rock budget with gray-area disputes; victims then lack rocks for the next scam.

### 5.2 Asymmetric binding

Buyers can hurt **listings** (search visibility). Sellers can hurt **buyer profiles** (user-bound rocks) without moving the ad’s `rock_count`. That asymmetry matches “report the ad” vs “warn others about this inquirer,” but it also lets a dishonest seller smear a cautious buyer while the listing stays searchable if no inquirer rocks stick.

### 5.3 Publicity vs privacy

Throwing a rock flips `CanViewConversation` for **any authenticated user**: the thread is no longer private-to-participants. Rock-icon UX prefers showing bystanders the **opinion modal**, but the conversation modal builder still loads the full journal for viewers who can open the conversation. Design intent is “public dispute.” User expectation from years of private marketplace chat may be “report, but keep my DMs private.” Those conflict. Anything typed before the rock—meetup addresses, partial VIN chatter, insults—may become visible to a wider authenticated audience.

### 5.4 Thrower reclaim (unthrow)

Only the thrower can unthrow. FAQ/Terms describe reclaiming the rock after resolution. There is no accused-party unilateral “return.”

### 5.5 Soft threshold before hard exclusion

Two rocks still leave an ad in search. A sophisticated bad actor can absorb a couple of rocks (or pressure throwers to unthrow) and remain discoverable. Exclusion at `> MaxRockCount` is strong only if enough distinct inquirers spend scarce rocks and leave them outstanding.

### 5.6 AI opinion brittleness

Grok leaves the trust boundary (also noted in the phone-identity paper). Opinions can be wrong, culturally biased, or gamed by parties who write for the model. Regex redaction is best-effort. Image attachment improves visual-policy cases but increases cost/latency and still depends on model judgment. Cache means a bad first stored opinion sticks until invalidation.

### 5.7 Journal and backup footprint

Rock events live in encrypted journals; opinions live in Postgres plaintext fields (paraphrase, but durable). Backups export rock opinion rows. Deleting an account force-unthrows, but historical journal lines and prior opinion text may already have been seen or backed up.

### 5.8 Cold-start and capture

In a thin market, a few networked accounts can dominate rock outcomes for a niche. Phone friction raises cost; it does not equalize power between a lone hobbyist and a shop with three SIMs.

### 5.9 Structured reasons (partial)

Throws require a canned reason (`policy`, `conduct`, `deal`) stored in the journal. That helps humans and the LLM without a full ticket taxonomy. Free-text reasons are intentionally omitted.

---

## 6. Abuse and exploit narratives

Illustrative threat stories grounded in how the system works. Not claims of active CVEs.

### Narrative A — Rock brigading a competitor

A seller verifies three phone numbers, opens conversations on a rival’s listing (or uses friends), and throws three ad-bound rocks. The rival’s ad exits search once `rock_count` exceeds 2. The attacker’s cost is three phone-gated accounts and three committed rocks. Mitigation today: phone + Turnstile cost, outstanding-rock caps per account, visibility of who rocked (dispute trail). Residual: real humans can still brigading-attack a niche listing.

### Narrative B — Rock exhaustion / denial of reporting

Attacker engages a target in multiple gray disputes, inducing the target to throw rocks elsewhere or holding rocks against them until the target’s allotment is stuck. Target then cannot rock a clear scam. Scarcity becomes a DoS on the reporting tool itself.

### Narrative C — Seller smear without search impact

Owner throws rocks at inquirers who ask hard questions. Inquirer profiles show user-bound rocks; the owner’s ads stay at `rock_count = 0`. Bystanders may avoid messaging that seller’s counterparties—or may misread user rocks as “this buyer is bad” when the seller is the aggressor. Opinion scores may or may not correct the narrative.

### Narrative D — Privacy bait-and-switch

Parties negotiate in what feels like a private thread. One side throws a rock after collecting sensitive details. Authenticated bystanders can access the dispute; depending on entry path they may see opinion paraphrase or full journal content. The rock becomes a weaponized disclosure tool, not only a policy report.

### Narrative E — Opinion prompt injection / theater

A party pastes long “evidence” and legalistic language optimized to steer Grok’s 1–10 scale. Cached opinion then greets every rock-icon click. Until unthrow/invalidation, the theater is the public story.

### Narrative F — Ordinal / ID enumeration

Rock icons use ordinals; conversation IDs are integers. Authenticated attackers can probe public rock conversations and harvest dispute metadata, usernames, and opinion text—useful for harassment or for mapping who fights with whom in a local niche.

### Narrative G — Stale rocks after social resolution

Parties settle off-platform. Thrower forgets to unthrow (or refuses). Ad remains partially or fully suppressed; thrower’s rock stays locked. No seller-side “return” exists in code despite FAQ language.

### Narrative H — Complement with classic account takeover

SIM-swap or password stuffing (see phone-identity paper) yields an account that can throw or unthrow rocks, rewrite journals as a participant, and manipulate dispute optics. Rocks inherit whatever account-security posture the platform has.

---

## 7. Legal issues

This section is analysis for internal review, **not legal advice**. Jurisdiction and facts matter; Rocky Ads Terms already stake several positions.

### 7.1 Self-policing vs platform duties

Terms describe Rocky Ads as “self-policed with rocks” and tell users not to email for listing problems, copyright claims, or other content issues—use rocks—while reserving the right to remove content or restrict accounts for law or service protection.

Tension: some regimes expect a designated channel for illegal content, copyright (DMCA in the U.S.), or child-safety reports. Pushing everything through a peer rock UX can look like obstruction if it is the *only* path and staff never act on non-rock signals. The reserve-power sentence is load-bearing. Rocks should remain the **preferred UX for ordinary listing disputes**, not the exclusive lawful intake for categories where statute or policy demands operator action.

### 7.2 Intermediary liability (e.g., U.S. Section 230 framing)

Peer-visible rocks and search demotion are user-driven signals. Platform-authored **rock opinions** are different: the service generates and displays its own assessment. That is closer to platform speech / tools than to pure hosting of third-party flags. Whether that affects intermediary-liability analysis is a counsel question; product-wise, opinions should stay clearly labeled as provisional, non-binding, and not an official finding of fraud or legality.

### 7.3 Defamation and reputational harm

Public rock icons and user-bound rocks are reputation statements. False brigading can damage sellers; false owner-rocks can damage buyers. Terms shift disputes between users and limit liability, but they do not make defamation impossible. Design mitigations that help legally and socially: scarcity, phone cost, visible dispute context, opinion framing as paraphrase not verdict, ability to unthrow, and operator bans for abuse.

### 7.4 Privacy and publicity of messages

Making rocked threads viewable to other members is a material change in processing purpose relative to ordinary private messaging. Privacy Policy / Terms should stay aligned with that reality (they already point rocks as the report path and note rocks on ads are visible). Surprises—“I didn’t know reporting published our chat”—create regulatory and trust risk under unfair-surprise theories even when Terms mention visibility. Prefer explicit throw-time copy: throwing a rock makes this dispute visible to other members.

### 7.5 Intellectual property reports

A rock is not a DMCA-compliant notice (identity of complainant, signature, specific works, sworn statements, etc.). If Terms tell users to use rocks for copyright claims, counsel should confirm whether a parallel statutory intake exists or whether Terms should carve IP claims out of the “don’t email / only rocks” rule. Mixing IP takedown into peer scarcity mechanics is a poor fit.

### 7.6 Illegal and high-risk content

Rocks are too slow and too peer-dependent for CSAM, imminent threats, and clear illegal goods. Operator tools and mandatory reporting obligations (where applicable) must bypass rock thresholds. Search exclusion after three rocks is irrelevant if the listing should never have been up.

### 7.7 Consumer protection and “assumed good”

Marketing that ads are assumed good until rocked is accurate as a product default; it must not be read as a guarantee of safety. Disclaimers already stress AS IS service and user–user dispute limits. Keep rock copy descriptive (“peers marked a dispute”) rather than certifying (“verified safe” / “cleared by Rocky”).

### 7.8 AI assessments

Automated fault scoring raises accuracy, bias, and transparency issues. Prompt rules (no legal advice, Owner/Inquirer only, paraphrase) reduce some risk. Remaining issues: users treating a 1–10 score as platform judgment; cross-border data transfer to the model provider; retention of opinion text after parties wanted the fight forgotten. Invalidation on unthrow helps; backups and screenshots do not forget.

### 7.9 Unthrow authority

FAQ/Terms state the thrower can unthrow and reclaim the rock, matching `UnthrowRock`.

---

## 8. Risk matrix (condensed)

| Threat / issue | Primary impact | Severity if unmitigated | Notes |
|----------------|----------------|-------------------------|--------|
| Brigading with phone-gated sockpuppets | Search exclusion of target ads | Medium–High | Costly but feasible in niches |
| Rock-budget exhaustion | Cannot report real scams | Medium | Scarcity DoS |
| Owner smear (user-bound rocks) | Buyer reputation harm | Medium | No ad `rock_count` change |
| Publicity of private journal | Privacy / harassment | High | By design when rocked; throw confirm warns |
| Opinion injection / wrong AI | Misleading bystanders | Medium | Cache amplifies; preview may reduce surprise throws |
| Relying on rocks for CSAM/IP statutory duties | Legal exposure | Critical if exclusive | Keep operator path |
| Account takeover → rock manipulation | Trust-signal integrity | High | Inherited from auth |
| Thin-market capture | Local monopoly on narrative | Medium | Cold start |

---

## 9. Recommendations aligned with the rocks philosophy

Status relative to the current codebase and docs:

1. ~~**Align unthrow policy with public copy.**~~ **Done.** FAQ/Terms say the thrower unthrows.  
2. ~~**Throw-time consent copy.**~~ **Done.** Confirm modal covers publicity, scarcity, unthrow, provisional assessment.  
3. **Clarify bystander surface.** Decide deliberately: opinion-only for non-participants everywhere (including conversation ID entry), vs full transcript transparency. Document the choice in Terms.  
4. **Carve statutory intakes.** Keep rocks for ordinary listing disputes; publish how CSAM, imminent harm, and copyright notices reach operators without requiring rock scarcity. Soften “do not email… copyright claims” if no compliant alternate exists.  
5. ~~**Label opinions harder.**~~ **Done** on preview panel; keep labels on the public opinion modal too.  
6. ~~**Consider light structure.**~~ **Done.** Canned reasons `policy` / `conduct` / `deal` in journal + prompts.  
7. **Watch brigading metrics.** Spike detection on multi-account rocks against one ad in a short window; do not silently auto-pardon (that would reintroduce opaque moderation), but give operators a signal.  
8. **Preserve scarcity.** Resist “unlimited rocks” feature requests; abundance recreates flag spam. If raising caps, raise phone/friction cost in tandem.  
9. **Invalidate and retention.** Confirm backup/restore and account-deletion paths for `rock_opinions`; document retention. Image confirms already invalidate.  
10. **Keep innocent-until-rocked.** Do not bolt on pay-to-verify badges that recreate the marketplace theater rocks were meant to avoid—unless a separate, honest product line needs them.  
11. ~~**Image-aware assessments.**~~ **Done** (up to 4× `480w` via multimodal Grok).

---

## 10. Conclusion

Rocks are Rocky Ads’ answer to a question most marketplaces dodge: **can peers enforce trust with a scarce, visible, conversation-bound signal instead of an infinite report button and a hidden queue?** The implementation is coherent: binding by thrower role, denormalized ad counts, search exclusion past two rocks, journaled throw/unthrow, and Grok opinions for bystanders.

The philosophy—assumed good, scarcity, dispute-as-conversation, publicity-as-enforcement—matches phone-minimal identity and the classifieds metaphor. The same philosophy creates the flaws that matter: gaming under scarcity, publicity of chats, asymmetric seller/buyer weapons, AI fallibility, and legal categories that cannot wait for three peers to spend rocks.

The correct reading is not “rocks replace trust & safety,” but **“rocks replace *casual* flag spam with costly peer disputes, while the operator retains narrow duties for law and platform integrity.”** That bet pays off only if copy matches code, throwers understand publicity, statutory intakes stay open, and rock abuse is treated as seriously as listing fraud itself.

---

## Appendix A — Primary code map

| Area | Location |
|------|----------|
| Throw / unthrow / counts / ordinals / reasons | `internal/rock/rock.go`, `internal/rock/reason.go` |
| HTTP handlers | `internal/handler/rock.go` (confirm, preview, throw), `AdRockConversationHandler` / `UserRockConversationHandler` |
| Routes | `cmd/server/main.go` (`/auth/ad/:id/rock/…`, `/auth/conversation/:id/rock/…`, `/auth/user/:id/rock/…`) |
| Conversation publicity / modal access | `internal/message/view.go` (`CanViewConversation`), `internal/message/message.go` |
| Journal rock events | `internal/journal/journal.go`, `internal/message/journal.go` |
| Rock opinions | `internal/rockopinion/` (prompt, redact, cache, images, preview) |
| UI | `internal/ui/rock.go`, `internal/ui/rock_confirm.go`, `internal/ui/rock_opinion.go` |
| Grok multimodal | `internal/service/grok/grok.go` |
| Search filter | `internal/search/metadata.go` (`rock_count <= MaxRockCount`) |
| Config knobs | `internal/config/config.go` (`MaxRockCount`, `MaxOutstandingRocks`) |
| Schema | `internal/db/schema.sql` (`conversations.rock_*`, `rock_opinions`) |
| FAQ / Terms / Privacy | `internal/ui/faq.go`, `internal/ui/legal.go` |
| Lifecycle cleanup | `internal/user/user.go`, ad delete/expire paths calling `UnthrowActiveForAd` |
| Embedding context | `internal/vector/ad.go` (`rockContext`) |

## Appendix B — Related public claims

In-app FAQ (`/faq/rocks`): three rocks at join; throw on policy/problem ads; choose a reason and optionally review assessment; starts a conversation; ads with more than `MaxRockCount` rocks leave search; thrower can unthrow and reclaim; use wisely.  
Terms (“Reporting and removal”): self-policed with rocks; rocks visible; search exclusion; reason + assessment before throw; thrower unthrows; do not email for listing/copyright/content issues—use rocks; operator may still act.  
Privacy (“Content reports”): listing problems handled in-product via rocks.  
Local launch doc: rocks mark bad actors; ads assumed good until rocked; do not expect rocks to substitute for supply.

## Appendix C — Parameter cheat sheet

| Constant | Value (Aug 2026) | Meaning |
|----------|------------------|---------|
| `MaxOutstandingRocks` | 3 | Max rocks a user may have thrown at once; granted at signup |
| `MaxRockCount` | 2 | Max `ads.rock_count` still allowed in search (`<= 2`); exclusion when count exceeds this |
