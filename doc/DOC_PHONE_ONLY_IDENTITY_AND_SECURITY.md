# Phone-Only Identity at Rocky Ads

## Privacy by Minimal Collection, and the Security Questions That Remain

**Document type:** Design and security analysis  
**Audience:** Product, engineering, and security reviewers  
**Scope:** How Rocky Ads uses a phone number as the sole contact identifier for registration, account management, and communications; why other personal data is intentionally omitted; and how the system can still be attacked despite a small PII footprint.  
**Based on:** Current Rocky Ads site implementation (Go / Fiber / Postgres / Twilio / JWT), August 2026.

---

## Abstract

Rocky Ads is classified advertising rebuilt for the web: post an ad, get contacted through in-app messaging, and keep your phone number hidden from other users. The product thesis is deliberately austere: **no email, no social graph, no payment card on file for posting, and no profile dossier.** The one piece of personal contact information the system requires is a text-capable phone number in E.164 form.

That choice is not a gimmick. It is a privacy architecture. Collecting less personal data reduces the blast radius of a breach, simplifies compliance narratives, and matches the mental model of classic newspaper classifieds—where a phone number was enough to complete a transaction. At the same time, **low PII is not the same as low risk.** Phone numbers are high-value identifiers. SMS channels can be abused. Sessions, passwords, message content, uploaded media, and operational secrets remain attractive targets. This paper explains the product rationale, maps how phone numbers are used in the running system, and explores realistic exploit paths that survive even when the user record is intentionally thin.

---

## 1. Product thesis: one identifier, many jobs

Traditional consumer marketplaces accumulate identity over time: email for login and receipts, phone for two-factor auth, address for shipping, payment methods for fees, and often a social login for convenience. Each field expands the attack surface and the privacy liability.

Rocky Ads reverses that default. The FAQ states the ideal plainly: the service would prefer to collect no personal information at all, but still needs a reliable way to reach a real person. A text-capable phone number fills three roles at once:

1. **Proof of a reachable human** during registration (one-time SMS verification).
2. **Account recovery channel** when a user loses password access (inbound SMS proof of possession).
3. **Operational notification channel** when unread messages arrive (outbound SMS with a link into the authenticated messages UI).

Everything else is optional or absent by design. Login itself is username plus password, not phone-as-password. Other users never see the number. Marketing SMS unrelated to the account is out of scope. The public surface of a member is closer to a classified-ad persona than to a social-network profile: username, ads, optional account picture, and activity signals—not a contact card.

The economic and UX claims in the product README reinforce the same boundary: no Facebook friends graph, no posting fees, no credit cards required to participate. Phone verification substitutes for the weak “anyone can invent an email” barrier without importing an entire CRM stack.

---

## 2. What counts as PII here—and what Rocky Ads refuses to collect

For this paper, **PII** means data that identifies or can reasonably identify a natural person. Phone numbers clearly qualify. Usernames may or may not, depending on whether they encode a real name; Rocky Ads encrypts them at rest anyway. Content users voluntarily type into ads or messages can contain arbitrary sensitive data; that is user-generated content, not a registration requirement.

### Intentionally not collected as account fields

| Category | Status |
|----------|--------|
| Email address | Not collected |
| Legal / government ID | Not collected |
| Mailing / home address | Not collected (ads may have location facets) |
| Payment card / bank account | Not required for posting |
| Social login / friends graph | Not used |
| Date of birth, gender, employer | Not collected |
| Device advertising IDs as identity | Not used as the account key |

### What is stored for an account

At the user-record level, the durable personal contact field is the phone number (E.164), stored encrypted with AES-GCM under a server-held key, with a SHA-256 lookup hash for uniqueness checks. The username is similarly encrypted. Passwords are stored as Argon2id hashes. Soft-delete leaves a time-bounded hold so a phone cannot be immediately reused after account deletion.

Ephemeral or operational stores also touch the phone briefly: Twilio Verify holds OTPs out-of-process; recovery sessions are bound after an inbound text; SMS notification queue rows and server logs may include request metadata or phone identifiers during SMS flows. Administrators with elevated access can decrypt phones for support and moderation. Backups preserve ciphertext blobs.

The privacy claim is therefore precise: **Rocky Ads does not need other contact PII to operate the product**, not that the system stores zero information about people. Ads, images, message journals, click history, and geo facets still create a behavioral and content footprint. Minimizing *required contact PII* is the design win; it does not erase all privacy risk.

---

## 3. How the phone number is used end-to-end

### 3.1 Registration

Registration is a three-step HTMX flow:

1. **Step 1 — claim username and phone.** The user consents to SMS (`offers=true`), passes Turnstile (unless test registration), the system checks availability and rate limits, optionally screens the username via a moderation model, and starts a **Twilio Verify** SMS (Fraud Guard + Verify geo).
2. **Step 2 — prove possession.** The user submits the OTP. The server runs a Verify check; on success it sets a short-lived HttpOnly **`reg_ticket` cookie**: a JWT (HS256 with `JWT_SECRET`) bound to that username and phone (~10 minute TTL). The raw OTP is not placed in the password form.
3. **Step 3 — set password and accept terms.** Before `CreateUser`, the handler **consumes** the cookie (`registrationticket.Consume`): verify signature and expiry, check username/phone match the form, clear the cookie. Only then is the user row created, a JWT session cookie issued, and the browser redirected to the authenticated welcome path. Replay of a consumed flow is blocked in practice by username/phone uniqueness on create.

Test environments may allow a reserved phone range (`+1555010xxxx`) to skip Turnstile and Verify when explicitly enabled; those paths still issue and consume the registration cookie. That flag must never be on in production.

### 3.2 Day-to-day account management

Once registered, the phone number remains the user’s contact anchor:

- Settings displays the current number and supports **change phone** (password + OTP to the new number).
- Users can toggle SMS notifications (`sms_opted_out`).
- Password changes invalidate sessions by rotating a salt bound into the JWT claims.
- Account deletion is soft-delete with a phone reuse hold (on the order of ten days).
- Distance-unit defaults can be inferred from phone region without asking for a home address.

Login remains username/password. The phone is not a daily credential; it is the recovery and notification backbone.

### 3.3 Communications

Outbound SMS covers verification codes and unread-message alerts. A background worker drains a notification queue, applies suppression (for example, avoiding SMS spam when the user was recently texted or has no unread messages), decrypts the phone, and sends via Twilio. Inbound SMS on the Twilio number handles carrier-style `STOP` handling for pending verification traffic and `RECOVER <code>` for account recovery.

Account recovery is possession-based rather than email-based: the browser starts a short-lived recovery session and shows a code; the user texts that code from their registered number; the webhook, after Twilio signature verification, binds the session to the matching user; the browser then reveals the username and allows password reset.

### 3.4 What other users see

Other members interact through usernames, ads, and the messaging product. The phone stays server-side. That preserves the classifieds metaphor—people call *the system’s messaging layer*, not a publicly printed mobile number—while still letting Rocky Ads nudge the seller when activity happens.

---

## 4. Why “phone only” is a real privacy improvement

### 4.1 Smaller breach catalog

If an attacker obtains a user table, the absence of emails, payment instruments, and home addresses means there is less immediately monetizable identity material of those types. There is no email list to dump into spam corpora, no card PANs to sell, no mailing list to dox. For many breach narratives, that is a meaningful reduction in harm categories.

### 4.2 Smaller third-party fan-out

Email-centric products often share addresses with ESP vendors, support desks, analytics, and CRM tools. Rocky Ads’ contact channel is SMS via Twilio. That is still a third party—and a concentrated one—but it avoids the sprawling email marketing stack that usually accumulates shadow copies of identity.

### 4.3 Better match to threat of over-collection

Many privacy failures are not exotic zero-days; they are products that asked for too much “just in case.” Phone-only registration is a forcing function against that habit. Features that would normally demand email (receipts, newsletters, magic links) must either be redesigned around SMS/in-app UX or rejected.

### 4.4 Human-cost reduction for casual users

Classifieds users often want transactional anonymity relative to strangers. Hiding the number from other users while still verifying it for the platform is a concrete improvement over printing a cell number in an ad.

---

## 5. Security properties that already support the model

Rocky Ads is not relying on obscurity alone. Relevant controls in the current stack include:

- **Password hashing** with Argon2id.
- **Field encryption** for username and phone (AES-GCM with per-user key derivation from `DB_ENCRYPTION_KEY`).
- **Message journal encryption** at rest.
- **JWT session cookies** (HttpOnly, SameSite=Strict, Secure outside local development), with password-salt binding so password changes invalidate tokens.
- **CSRF** protection (double-submit), with a deliberate exemption for the Twilio webhook path.
- **Helmet / CSP** and related browser hardening headers.
- **Rate limits** on registration and recovery starts, plus a global per-IP ceiling.
- **OTP attempt limits** and expiry (Twilio Verify).
- **Twilio Verify** for register/change-phone OTP, then a signed HttpOnly **`reg_ticket` cookie** (JWT) for step 3 so the password form never round-trips the raw OTP; a parallel client without that cookie cannot finish signup.
- **Cloudflare Turnstile** and per-phone OTP start limits before Verify create.
- **Recovery secrets** stored as HMAC digests rather than raw tokens.
- **Twilio request signature verification** on inbound webhooks.
- **Soft-delete phone hold** to slow account recycling abuse.

These controls matter because the phone-only model concentrates trust in SMS possession, password strength, session integrity, and the secrecy of encryption/JWT/Twilio credentials. The following sections assume those controls exist and ask what still goes wrong.

---

## 6. Low PII is not low attack surface

A useful mental model: **privacy minimization shrinks the value of some data dumps; it does not shrink the number of ways to hijack an account, spam a channel, or extract user-generated content.** In some cases it even intensifies pressure on the remaining identifier.

### 6.1 The phone number is a high-value singleton

On the **carrier / SMS-possession** axis, Rocky Ads is in the same club as WhatsApp (and Signal, iMessage registration, and many “phone-only” messengers): whoever can receive SMS for the number can prove possession to the service. Rocky Ads does not escape SIM swap, SS7 abuse, malicious MVNO apps, shared family plans, stolen handsets, or social-engineered ports. That weakness lives at the phone network, not in how little other PII is stored.

Where Rocky Ads **differs** from WhatsApp is how the number is *used* after proof:

| | WhatsApp-style messenger | Rocky Ads |
|--|--------------------------|-----------|
| Daily auth | Phone (re)registration / device link | Username + password |
| Public / peer identity | Number is the account handle; contacts can discover you | Number is hidden; peers see a username |
| How strangers reach you | Message the number (or a linked identity) | In-app messaging to the username / ad |
| Phone’s jobs here | Identity + delivery address | Verify, recover, and notify only |

So “phone only” means different things. WhatsApp makes the number the *social* identity. Rocky Ads makes it a *platform-side contact key* and deliberately keeps a second channel (password) for day-to-day login. That split helps in one direction and hurts in another:

- **Helps:** a SIM swap alone does not equal an already-logged-in browser session the way a fresh WhatsApp registration on a stolen number can. The attacker still needs recovery (SMS) *or* the password (stuffing / phishing).
- **Hurts:** password stuffing and weak passwords become first-class risks WhatsApp largely sidesteps by not using passwords. And for **registration, change-phone, and recovery**, SMS possession remains the lifeline—closer to WhatsApp’s model than to “email + SMS 2FA,” where those are separate channels.

**Implication:** privacy minimization does not buy Rocky Ads a stronger phone channel than WhatsApp’s. It buys a thinner account dossier and a hidden number. Carrier-layer SMS identity risk still applies wherever the product gates on SMS.

### 6.2 Registration and OTP economics

Even with IP rate limits, SMS OTP systems face familiar abuses:

- **SMS pumping / toll fraud:** bots register with numbers that route SMS to premium or incentivized destinations, burning Twilio spend. Mitigations: Twilio Verify + Fraud Guard, Verify geo US+CA allow (elsewhere monitor+block fraud), Messaging geo US+CA only, Turnstile, per-IP/per-phone OTP start limits, spend alerts — see [DOC_SMS_OTP_AND_PUMPING_DEFENSES.md](DOC_SMS_OTP_AND_PUMPING_DEFENSES.md).
- **OTP flooding:** harassment by repeatedly triggering codes to a victim phone (limited by registration rate limits, Turnstile, per-phone OTP start limits, and Twilio Verify Fraud Guard; change-phone shares the same OTP-start limits).
- **Step binding (mitigated):** registration previously trusted step 2 alone at password submission. That gap is closed: step 2 completes Twilio Verify and sets a signed `reg_ticket` cookie; step 3 must present that cookie matching the form. See [DOC_SMS_OTP_AND_PUMPING_DEFENSES.md](DOC_SMS_OTP_AND_PUMPING_DEFENSES.md).

OTP codes for register/change-phone are owned by Twilio Verify and are not stored in the application database.

### 6.3 Password and session attacks still dominate day-to-day risk

Because daily login is username/password:

- credential stuffing and password spraying remain relevant;
- XSS that steals nothing “PII-like” can still steal a session cookie if script injection ever appears;
- CSRF protections must stay correct for state-changing routes;
- a leaked `JWT_SECRET` is catastrophic regardless of how little PII is stored.

Low PII does not help a user whose password is `Summer2024!` and whose username is guessable from their public ads.

### 6.4 Account recovery as a target

The recovery design is elegant for a no-email product, but it creates a crisp attack graph:

1. Attacker starts many recovery sessions (mitigated by per-IP rate limits).
2. Attacker tries to guess or intercept the on-screen code path, or socially engineer the victim to text a code.
3. If the attacker can send SMS *from* the victim’s number (compromised phone, spoofing depending on carrier path—Twilio’s `From` is the authentic inbound peer for webhook handling), they bind recovery and reset the password.

Webhook security is load-bearing. Signature verification must use the correct public URL; trusting mis-set `X-Forwarded-*` headers in reverse-proxy configurations can weaken or break that validation. The recovery endpoint is CSRF-exempt by necessity for Twilio, which makes signature checks non-optional.

### 6.5 Lookup hashes and offline correlation

Encrypting phone and name at rest is strong against casual database reads **if** `DB_ENCRYPTION_KEY` remains secret. Unsalted (or un-peppered) SHA-256 hashes of E.164 phones are still amenable to dictionary attack: the phone number space is structured and far smaller than a general password space. An attacker with the hash column can test candidate numbers offline and correlate users to real-world identities even without the AES key. That does not expose message plaintext by itself, but it undercuts the narrative that ciphertext alone equals anonymity.

### 6.6 Content and metadata are still PII-rich

Users will put phone numbers, meetup addresses, and workplace details into ad text and messages. Image uploads may contain faces, license plates, or document photos. Geo facets and locations tables describe where goods are. Click and bookmark tables describe interest graphs. Encrypted journals protect disks and backups better than plaintext, but participants and anyone who can act as a participant (account takeover) can read them.

In other words: **the registration form is minimal; the product corpus is not.** A breach of message stores or object storage is a privacy incident even if the `users` table lacks email.

### 6.7 Admin and insider risk

A thin user record still decrypts to a phone number for admins. Compromised admin JWTs, over-broad `is_admin` claims that survive until token expiry after demotion, or jump-server access to production keys turn the “hidden phone” property into an internal disclosure problem. Low external PII does not remove privileged-path risk.

### 6.8 Third-party and configuration failures

| Dependency | Failure mode |
|------------|----------------|
| Twilio | Account takeover → SMS redirect, outbound spam as Rocky Ads, inbound webhook forgery if signatures fail |
| `TWILIO_WEBHOOK_URL` misconfig | Notification SMS can point users at a phishing host that mimics login |
| Grok username screening | Usernames leave the trust boundary to a model API |
| MinIO / presigned URLs | Stolen long-lived GET URLs leak images; PUT URL abuse if minting is too loose |
| `ALLOW_TEST_REGISTRATION` | Skips real SMS proof for allowlisted numbers if enabled in prod |
| Ollama / other internal services | Usually not PII routers, but expand operational attack surface |

### 6.9 Product abuse that is not “hacking the crypto”

Even with perfect cryptography, a classifieds site faces:

- scam ads and advance-fee fraud,
- spam accounts (phone verification raises cost but does not eliminate VoIP/SMS farms),
- harassment via messaging (SMS notifications become part of the harassment loop if not rate-limited and easy to disable),
- enumeration side channels (phone-hold messages, timing of availability checks, distinct error paths).

These harms do not require a large PII database. They require a working marketplace.

---

## 7. Example exploit narratives (even with phone-only accounts)

The following scenarios are illustrative threat stories grounded in how the system works. They are not claims of currently exploited CVEs.

### Narrative A — SIM swap recovery takeover

An attacker social-engineers the victim’s carrier, ports the number, starts recovery on Rocky Ads, texts `RECOVER <code>` from the newly controlled line, resets the password, and reads message history about local meetups. No email inbox was needed. The “low PII” database still yielded conversation content and ad activity once the account was owned.

### Narrative B — Database backup without the encryption key—but with hashes

An attacker steals a logical backup. AES-GCM phone ciphertext resists direct read. Register/change-phone OTPs are not in the app database (Twilio Verify). Unsalted phone hashes still enable bulk correlation of which real numbers have accounts. The attacker may then phishing-SMS those numbers with a lookalike “verify your Rocky Ads phone” link.

### Narrative C — Registration step confusion / race *(mitigated)*

**Previously:** an attacker who knew an in-flight username + phone could POST step 3 without the OTP, because account creation trusted that step 2 had already succeeded. The victim’s verified phone could become an ownership-transfer mechanism.

**Now:** step 2 completes Twilio Verify and sets a signed `reg_ticket` cookie; step 3 verifies and clears that cookie. Without the cookie, enrollment fails. Remaining related risks are cookie theft from the browser and SMS/Verify-channel attacks—not username/phone guessing alone. See [DOC_SMS_OTP_AND_PUMPING_DEFENSES.md](DOC_SMS_OTP_AND_PUMPING_DEFENSES.md).

### Narrative D — SMS brand phishing via config foot-gun

A deployment sets `TWILIO_WEBHOOK_URL` to a wrong host. Legitimate unread-message texts contain that host. Users, trained by Rocky Ads to trust SMS links into `/auth/user/messages`, submit passwords to an attacker site. The application’s cryptography never failed; the operational URL did.

### Narrative E — Credential stuffing without any phone involvement

Public usernames from ads are tried with passwords from unrelated breaches against `/api/login`. Without a dedicated login rate limit beyond the global ceiling, spraying may be practical. On success, the attacker disables SMS notifications or changes the phone after password entry, isolating the victim from alerts. Phone-only registration did not slow this path because login does not require the phone.

### Narrative F — Marketplace trust exploits

A scammer verifies a disposable SMS number, posts attractive ads, and social-engineers buyers inside encrypted message journals. When reports arrive, the operator has a phone hash and ciphertext—not a rich KYC file. The low PII posture that protects honest users also limits investigative breadcrumbs.

---

## 8. Risk matrix (condensed)

| Threat | Needs large PII store? | Primary impact | Severity if unmitigated |
|--------|------------------------|----------------|-------------------------|
| SIM swap / SMS interception | No | Account recovery / phone change | High |
| Password stuffing / weak passwords | No | Full account takeover | High |
| JWT or encryption key leak | No | Mass session or field decrypt | Critical |
| Twilio console compromise | No | SMS redirect, spam, trust break | Critical |
| Presigned media URL leak | No | Image/PII-in-photo exposure | Medium |
| Admin token abuse | No | Bulk phone disclosure | High |
| SMS pumping | No | Direct financial loss | Medium–High |
| Content scams / harassment | No | User harm, brand trust | High |
| Offline phone-hash correlation | No | Re-identification of members | Medium |

The pattern is consistent: **almost none of the serious threats require Rocky Ads to have collected email or credit cards.**

---

## 9. Recommendations aligned with the phone-only philosophy

These suggestions aim to harden the model without abandoning it.

1. ~~**Treat registration as a single server-side capability token.**~~ **Done.** Step 2 completes Twilio Verify and sets a signed HttpOnly `reg_ticket` JWT cookie; step 3 verifies binding and clears it. OTP codes are not stored in-app.
2. ~~**Hash or encrypt OTPs at rest.**~~ **Done for register/change-phone:** OTP codes are owned by Twilio Verify (not stored in the app database).
3. **Pepper phone/name lookup hashes** with a server-side secret so offline dictionary attacks on E.164 space require both DB and secret compromise.
4. **Add login-specific throttling and lockouts** (per username and per IP), not only global request ceilings.
5. ~~**Rate-limit all SMS-sending endpoints** and monitor Twilio spend.~~ **Done (OTP path):** Twilio Verify + Fraud Guard, Turnstile, per-phone OTP start limits; spend-alert runbook in [DOC_SMS_OTP_AND_PUMPING_DEFENSES.md](DOC_SMS_OTP_AND_PUMPING_DEFENSES.md). Notification SMS remain on Programmable Messaging.
6. **Make STOP and in-app opt-out consistent** so carrier unsubscribe language matches `sms_opted_out` behavior users expect.
7. **Bound notification link bases** to a first-party canonical site URL distinct from misconfigurable webhook bases where possible; alert on host mismatch.
8. **Shorten admin phone exposure** (just-in-time decrypt, audit logs) and invalidate admin privilege changes immediately rather than waiting on JWT lifetime alone.
9. **Document SIM-swap reality** in user-facing recovery copy: phone possession is powerful; users should protect carrier accounts.
10. **Keep `ALLOW_TEST_REGISTRATION` impossible in production** via startup refusal when release mode is live.
11. **Preserve the non-collection stance** when new features are proposed—receipts, digests, and “magic links” should not silently reintroduce email as a second identity without a deliberate privacy review.

---

## 10. Conclusion

Rocky Ads’ phone-only approach is a coherent privacy strategy: one practical contact channel, hidden from other users, reused for verification, recovery, and message notifications, with a deliberate refusal to assemble an email-and-profile dossier. That stance reduces certain breach harms and keeps the product honest about classifieds-scale identity.

It does not, however, create a small security problem space. Concentrating identity into SMS possession raises the stakes of carrier-side attacks and OTP plumbing. Passwords, JWTs, webhook signatures, encryption keys, admin tools, media storage, and user-generated content remain full-fidelity targets. The correct reading of the architecture is not “little PII, therefore safe,” but **“little PII, therefore the remaining controls on phone proof, sessions, secrets, and content access must be excellent.”**

Privacy minimization is necessary and valuable. For Rocky Ads, it is also a bet that engineering discipline around a single contact identifier can outperform the false comfort of collecting more personal data “for security.” That bet only pays off if the phone channel, the password channel, and the operational envelope are hardened with the same seriousness that motivated collecting so little in the first place.

---

## Appendix A — Primary code map

| Area | Location |
|------|----------|
| Registration / OTP handlers | `internal/handler/register.go` (Verify + `reg_ticket`) |
| Twilio Verify | `internal/service/verify/` |
| Turnstile | `internal/service/turnstile/` |
| Registration ticket cookie | `internal/registrationticket/`, `internal/cookie/register_ticket.go` |
| Login / JWT | `internal/handler/login.go`, `internal/handler/jwt.go`, `internal/cookie/jwt.go` |
| Recovery | `internal/handler/recover.go`, `internal/accountrecovery/` |
| SMS send/queue/worker/webhook | `internal/service/sms/`, `internal/handler/sms.go` |
| Settings / change phone | `internal/handler/user.go`, `internal/user/` |
| Encryption helpers | `internal/encryption/`, `internal/user/encryption.go` |
| Schema | `internal/db/schema.sql` |
| Privacy / FAQ copy | `internal/ui/legal.go`, `internal/ui/faq.go` |
| Env and dependency overview | `README.md`, `doc/DOC_TECH_STACK.md` |

## Appendix B — Related public claims

Product README: phone number is enough to get started; no email; number stays hidden from the classifieds-style public view.  
In-app FAQ (`/faq/phone-number`): phone is collected mainly for message notifications and reachability; email, mailing address, and payment details are not requested as contact profile data; numbers are not sold or shown to other users.
