import { useState, useRef, useEffect, useCallback } from 'react'
import './Deadlock.css'

// ── DATA ─────────────────────────────────────────────────────────────────────

const GAME_MODIFIERS = [
  { name: 'Everyone Bans Their Main', desc: "Before picking, each player calls out their main — that hero is off limits for them this game." },
  { name: 'No Speaking English',    desc: 'All comms must be in another language. Speaking English = drink.' },
  { name: 'No Swearing',            desc: 'Keep it clean. Any swear word = drink.' },
  { name: 'Callouts Must Be Sung',  desc: 'Every in-game callout must be sung. Speaking normally = drink.' },
  { name: 'Dead Volume Off',        desc: 'When you die, turn your volume off until your next death.' },
  { name: '3rd Person Only',        desc: 'No "I" or "me". Everyone speaks in 3rd person all game.' },
  { name: 'No Dead Chat',           desc: "When you're dead: no speaking, no pinging. Any violation = drink." },
  { name: 'Lord & Squire',          desc: 'Two players are paired. The Squire helps the Lord secure kills and farm. Lords give orders, Squires follow. Roles locked all game.' },
  { name: 'Class Warfare',          desc: "Roll a dice: one player builds full Melee (Hunter's Aura, Crippling Headshot, etc.), one Gun only, one Spirit only, one Vitality only." },
]

const ROLES = ['Tank', 'Spirit', 'Gun', 'Support'] as const
const ABILITY_MAXES = [1, 2, 3, 4] as const

const ITEMS = {
  weapon: [
    'Close Quarters', 'Extended Magazine', 'Headshot Booster', 'High-Velocity Rounds',
    'Monster Rounds', 'Rapid Rounds', 'Restorative Shot', 'Active Reload', 'Stalker',
    'Fleetfoot', 'Intensifying Magazine', 'Kinetic Dash', 'Long Range', 'Melee Charge',
    'Mystic Shot', 'Opening Rounds', 'Slowing Bullets', 'Spirit Shredder Bullets', 'Split Shot',
    'Swift Striker', 'Titanic Magazine', 'Weakening Headshot', 'Alchemical Fire', 'Berserker',
    'Blood Tribute', 'Burst Fire', 'Cultist Sacrifice', 'Escalating Resilience', 'Express Shot',
    'Headhunter', 'Heroic Aura', 'Hollow Point', "Hunter's Aura", 'Point Blank', 'Sharpshooter',
    'Spirit Rend', 'Tesla Bullets', 'Toxic Bullets', 'Weighted Shots', 'Ballistic Enchantment',
    'Armor Piercing Rounds', 'Capacitor', 'Crippling Headshot', 'Frenzy', 'Glass Cannon',
    'Lucky Shot', 'Ricochet', 'Shadow Weave', 'Silencer', 'Spellslinger',
    'Spiritual Overflow', 'Crushing Fists',
  ],
  spirit: [
    'Extra Charge', 'Extra Spirit', 'Mystic Burst', 'Mystic Expansion', 'Mystic Regeneration',
    'Rusted Barrel', 'Spirit Strike', 'Golden Goose Egg', 'Arcane Surge', 'Bullet Resist Shredder',
    'Cold Front', 'Compress Cooldown', 'Duration Extender', 'Improved Spirit', 'Mystic Slow',
    'Mystic Vulnerability', 'Quicksilver Reload', 'Slowing Hex', 'Spirit Sap', 'Suppressor',
    'Decay', 'Disarming Hex', 'Greater Expansion', 'Knockdown', 'Rapid Recharge',
    'Silence Wave', 'Spirit Snatch', 'Superior Cooldown', 'Superior Duration',
    'Surge of Power', 'Tankbuster', 'Torment Pulse', 'Radiant Regeneration',
    'Arctic Blast', 'Boundless Spirit', 'Cursed Relic', 'Echo Shard', 'Escalating Exposure',
    'Ethereal Shift', 'Focus Lens', 'Lightning Scroll', 'Magic Carpet', 'Mercurial Magnum',
    'Mystic Reverb', 'Refresher', 'Scourge', 'Spirit Burn', 'Vortex Web', 'Transcendent Cooldown',
  ],
  vitality: [
    'Extra Health', 'Extra Regen', 'Extra Stamina', 'Healing Rite', 'Melee Lifesteal',
    'Rebuttal', 'Sprint Boots', 'Battle Vest', 'Bullet Lifesteal', 'Debuff Reducer',
    "Enchanter's Emblem", 'Enduring Speed', 'Guardian Ward', 'Healbane', 'Healing Booster',
    'Weapon Shielding', 'Reactive Barrier', 'Restorative Locket', 'Return Fire',
    'Spirit Lifesteal', 'Spirit Shielding', 'Bullet Resilience', 'Counterspell', 'Dispel Magic',
    'Fortitude', 'Fury Trance', 'Lifestrike', 'Majestic Leap', 'Metal Skin', 'Rescue Beam',
    'Spirit Resilience', 'Stamina Mastery', 'Trophy Collector', 'Veil Walker', 'Warp Stone',
    'Healing Nova', 'Cheat Death', 'Colossus', 'Divine Barrier', "Diviner's Kevlar",
    'Healing Tempo', 'Infuser', 'Inhibitor', 'Juggernaut', 'Leech', 'Phantom Strike',
    'Plated Armor', 'Siphon Bullets', 'Spellbreaker', 'Unstoppable', 'Vampiric Burst', 'Witchmail',
  ],
}

interface Hero {
  name: string
  img: string
  color: string
  drinkModifier: string
}

const HEROES: Hero[] = [
  { name: 'Abrams',     img: 'abrams.png',     color: '#E55B3C', drinkModifier: 'Pin 2+ people to wall and heavy melee both' },
  { name: 'Apollo',     img: 'apollo.png',     color: '#F0C040', drinkModifier: 'Ult double kill' },
  { name: 'Bebop',      img: 'bebop.png',      color: '#A855F7', drinkModifier: 'Every 50 stacks' },
  { name: 'Billy',      img: 'billy.png',      color: '#D97706', drinkModifier: 'Successful 3+ chain pull' },
  { name: 'Calico',     img: 'calico.png',     color: '#EC4899', drinkModifier: 'Kill (final blow) with a heavy melee' },
  { name: 'Celeste',    img: 'celeste.png',    color: '#3B82F6', drinkModifier: 'Killing blow with shield pop' },
  { name: 'Doorman',    img: 'doorman.png',    color: '#6B7280', drinkModifier: 'Cart into walker slam stun' },
  { name: 'Drifter',    img: 'drifter.png',    color: '#06B6D4', drinkModifier: 'Every 45 isolation stacks' },
  { name: 'Dynamo',     img: 'dynamo.png',     color: '#EAB308', drinkModifier: '4+ man ult' },
  { name: 'Graves',     img: 'graves.png',     color: '#9CA3AF', drinkModifier: 'Final blow with ult zombie' },
  { name: 'Gray Talon', img: 'grey_talon.png', color: '#A8956A', drinkModifier: 'Every 40 stacks' },
  { name: 'Haze',       img: 'haze.png',       color: '#8B5CF6', drinkModifier: 'Kill 4 people with one ult' },
  { name: 'Holliday',   img: 'holliday.png',   color: '#F59E0B', drinkModifier: 'Mid-air lasso leads to kill' },
  { name: 'Infernus',   img: 'infernus.png',   color: '#EF4444', drinkModifier: '4 man ult' },
  { name: 'Ivy',        img: 'ivy.png',        color: '#22C55E', drinkModifier: '3 people in stone form' },
  { name: 'Kelvin',     img: 'kelvin.png',     color: '#67E8F9', drinkModifier: '4 enemies in dome' },
  { name: 'Lady Geist', img: 'lady_geist.png', color: '#C084FC', drinkModifier: 'Kill with your 3' },
  { name: 'Lash',       img: 'lash.png',       color: '#F97316', drinkModifier: '4+ man ult' },
  { name: 'McGinnis',   img: 'mcginnis.png',   color: '#84CC16', drinkModifier: 'Trap 2 people in wall and kill them both' },
  { name: 'Mina',       img: 'mina.png',       color: '#DC2626', drinkModifier: 'Every 50 stacks' },
  { name: 'Mirage',     img: 'mirage.png',     color: '#0EA5E9', drinkModifier: 'Kill with heavy melee' },
  { name: 'Mo & Krill', img: 'mo_krill.png',   color: '#92400E', drinkModifier: 'Every 500 stacks' },
  { name: 'Paige',      img: 'paige.png',      color: '#10B981', drinkModifier: 'Heavy melee finishing kill' },
  { name: 'Paradox',    img: 'paradox.png',    color: '#6366F1', drinkModifier: '3+ man swap' },
  { name: 'Pocket',     img: 'pocket.png',     color: '#7C3AED', drinkModifier: '5+ man ult' },
  { name: 'Rem',        img: 'rem.png',        color: '#9CA3AF', drinkModifier: 'Every 4 enemy sinners successfully stolen' },
  { name: 'Seven',      img: 'seven.png',      color: '#FBBF24', drinkModifier: '4 man stun with ability 2' },
  { name: 'Shiv',       img: 'shiv.png',       color: '#34D399', drinkModifier: 'Kill 3 people in one ult' },
  { name: 'Silver',     img: 'silver.png',     color: '#94A3B8', drinkModifier: 'Killing blow on 3 people while in transform' },
  { name: 'Sinclair',   img: 'sinclair.png',   color: '#A78BFA', drinkModifier: 'Kill enemy (final blow) with their own ult' },
  { name: 'Venator',    img: 'venator.png',    color: '#F43F5E', drinkModifier: '360 killing blow with final shot' },
  { name: 'Victor',     img: 'victor.png',     color: '#34D399', drinkModifier: 'Stun 2 in ult' },
  { name: 'Vindicta',   img: 'vindicta.png',   color: '#7E22CE', drinkModifier: 'Every 40 stacks' },
  { name: 'Viscous',    img: 'viscous.png',    color: '#CA8A04', drinkModifier: 'Final blow 2 people in Ball' },
  { name: 'Vyper',      img: 'vyper.png',      color: '#16A34A', drinkModifier: 'Kill player while they are in stone' },
  { name: 'Warden',     img: 'warden.png',     color: '#3B82F6', drinkModifier: 'Kill 3 people while in ult' },
  { name: 'Wraith',     img: 'wraith.png',     color: '#C0A06B', drinkModifier: 'Heavy melee final blow' },
  { name: 'Yamato',     img: 'yamato.png',     color: '#B91C1C', drinkModifier: 'Land 4 heavy melees into kill on one person in a single fight' },
]

// ── HELPERS ──────────────────────────────────────────────────────────────────

function pickRandom<T>(arr: readonly T[]): T {
  return arr[Math.floor(Math.random() * arr.length)]
}

function pickOnePerCategory(exclude: string[] = []) {
  return (['weapon', 'spirit', 'vitality'] as const).map(cat => ({
    cat,
    item: pickRandom(ITEMS[cat].filter(i => !exclude.includes(i))),
  }))
}

function generateResult() {
  const cantBuy = pickOnePerCategory()
  const cantItems = cantBuy.map(c => c.item)
  const mustBuy = pickOnePerCategory(cantItems)
  return { role: pickRandom(ROLES), abilityMax: pickRandom(ABILITY_MAXES), cantBuy, mustBuy }
}

const CAT_COLORS: Record<string, string> = {
  weapon:   '#F97316',
  spirit:   '#A855F7',
  vitality: '#22C55E',
}

// ── GAME ROLL ─────────────────────────────────────────────────────────────────

// Build an infinitely-looping list for the drum effect
const DRUM_ITEMS = [...GAME_MODIFIERS, ...GAME_MODIFIERS, ...GAME_MODIFIERS]
const ITEM_H = 72 // px per drum row

function GameRoll() {
  const [phase, setPhase] = useState<'idle' | 'spinning' | 'done'>('idle')
  const [result, setResult] = useState<typeof GAME_MODIFIERS[0] | null>(null)
  const [drumOffset, setDrumOffset] = useState(0)
  const drumRef = useRef<HTMLDivElement>(null)
  const rafRef = useRef<number>(0)
  const spin = useCallback(() => {
    if (phase === 'spinning') return
    setPhase('spinning')
    setResult(null)

    const finalMod = pickRandom(GAME_MODIFIERS)
    const finalIdx = GAME_MODIFIERS.length + GAME_MODIFIERS.indexOf(finalMod)
    const targetOffset = finalIdx * ITEM_H

    // Start with high speed, ease to a stop
    const duration = 2200
    const startTime = performance.now()
    const startOffset = drumOffset

    function animate(now: number) {
      const elapsed = now - startTime
      const t = Math.min(elapsed / duration, 1)
      // Ease out cubic
      const eased = 1 - Math.pow(1 - t, 3)
      // Travel a long distance with wrapping feel
      const distance = targetOffset + GAME_MODIFIERS.length * ITEM_H * 2
      const current = startOffset + eased * distance

      setDrumOffset(current % (GAME_MODIFIERS.length * ITEM_H * 3))

      if (t < 1) {
        rafRef.current = requestAnimationFrame(animate)
      } else {
        setDrumOffset(finalIdx * ITEM_H)
        setResult(finalMod)
        setPhase('done')
      }
    }

    rafRef.current = requestAnimationFrame(animate)
  }, [phase, drumOffset])

  useEffect(() => () => cancelAnimationFrame(rafRef.current), [])

  // Visible drum rows: show 5 centered on current position
  const centerIdx = Math.round(drumOffset / ITEM_H)
  const visibleRows = [-2, -1, 0, 1, 2].map(offset => {
    const idx = ((centerIdx + offset) % DRUM_ITEMS.length + DRUM_ITEMS.length) % DRUM_ITEMS.length
    return { item: DRUM_ITEMS[idx], offset }
  })

  return (
    <div className="gr-page">
      {/* Decorative corner suits */}
      <div className="gr-suit gr-suit-tl">♠</div>
      <div className="gr-suit gr-suit-tr">♦</div>
      <div className="gr-suit gr-suit-bl">♣</div>
      <div className="gr-suit gr-suit-br">♥</div>

      <div className="gr-inner">
        {/* Header */}
        <div className="gr-heading">
          <div className="gr-dice">🎲</div>
          <div>
            <h2 className="gr-title">Strat Roulette</h2>
            <p className="gr-sub">1 per game &nbsp;·&nbsp; drink at your own pace</p>
          </div>
          <div className="gr-dice">🎲</div>
        </div>

        {/* Drum machine */}
        <div className="gr-drum-wrap">
          {/* Fade edges */}
          <div className="gr-drum-fade gr-drum-fade-top" />
          <div className="gr-drum-fade gr-drum-fade-bot" />
          {/* Center highlight */}
          <div className="gr-drum-highlight" />

          <div className="gr-drum" ref={drumRef}>
            {visibleRows.map(({ item, offset }) => (
              <div
                key={offset}
                className={`gr-drum-row ${offset === 0 ? 'gr-drum-row-active' : ''}`}
                style={{
                  '--row-dist': Math.abs(offset),
                  transform: `translateY(${offset * ITEM_H}px) scale(${offset === 0 ? 1 : 0.88 - Math.abs(offset) * 0.06})`,
                  opacity: offset === 0 ? 1 : 0.25 - Math.abs(offset) * 0.08,
                } as React.CSSProperties}
              >
                {item.name}
              </div>
            ))}
          </div>
        </div>

        {/* Roll button */}
        <button
          className={`gr-btn ${phase === 'spinning' ? 'gr-btn-spinning' : ''} ${phase === 'done' ? 'gr-btn-done' : ''}`}
          onClick={spin}
          disabled={phase === 'spinning'}
        >
          {phase === 'spinning' ? (
            <><span className="gr-btn-spinner" />Rolling…</>
          ) : phase === 'done' ? (
            'Roll Again'
          ) : (
            'Spin'
          )}
        </button>

        {/* Result card */}
        {result && (
          <div className="gr-result-card" key={result.name}>
            <div className="gr-result-glow" />
            <div className="gr-result-top">
              <span className="gr-result-badge">This Game</span>
              <h3 className="gr-result-name">{result.name}</h3>
            </div>
            <div className="gr-result-divider" />
            <p className="gr-result-desc">{result.desc}</p>
          </div>
        )}
      </div>
    </div>
  )
}

// ── HERO MODAL ─────────────────────────────────────────────────────────────────

interface HeroResult {
  role: string
  abilityMax: number
  cantBuy: { cat: string; item: string }[]
  mustBuy: { cat: string; item: string }[]
}

const CATS = ['weapon', 'spirit', 'vitality'] as const

// Display values shown in the card (either cycling or locked)
interface DisplayVals {
  role: string
  abilityMax: number | string
  cant: Record<string, string>
  must: Record<string, string>
}

function randomDisplay(): DisplayVals {
  return {
    role: pickRandom(ROLES),
    abilityMax: pickRandom(ABILITY_MAXES),
    cant: Object.fromEntries(CATS.map(c => [c, pickRandom(ITEMS[c])])),
    must: Object.fromEntries(CATS.map(c => [c, pickRandom(ITEMS[c])])),
  }
}

function resultToDisplay(r: HeroResult): DisplayVals {
  return {
    role: r.role,
    abilityMax: r.abilityMax,
    cant: Object.fromEntries(r.cantBuy.map(({ cat, item }) => [cat, item])),
    must: Object.fromEntries(r.mustBuy.map(({ cat, item }) => [cat, item])),
  }
}

function HeroModal({ hero, onClose }: { hero: Hero; onClose: () => void }) {
  const [rolling, setRolling]   = useState(true)
  const [display, setDisplay]   = useState<DisplayVals>(() => randomDisplay())
  const ivRef                   = useRef<ReturnType<typeof setInterval> | null>(null)
  const toRef                   = useRef<ReturnType<typeof setTimeout> | null>(null)

  function startRoll() {
    const final = generateResult()
    ivRef.current = setInterval(() => setDisplay(randomDisplay()), 80)
    toRef.current = setTimeout(() => {
      clearInterval(ivRef.current!)
      setDisplay(resultToDisplay(final))
      setRolling(false)
    }, 2000)
  }

  // Roll on mount
  useEffect(() => {
    startRoll()
    return () => { clearInterval(ivRef.current!); clearTimeout(toRef.current!) }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function reroll() {
    setRolling(true)
    startRoll()
  }

  useEffect(() => {
    function onKey(e: KeyboardEvent) { if (e.key === 'Escape' && !rolling) onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose, rolling])

  useEffect(() => {
    document.body.style.overflow = 'hidden'
    return () => { document.body.style.overflow = '' }
  }, [])

  return (
    <div className="hm-backdrop" onClick={!rolling ? onClose : undefined}>
      <div
        className="hm-modal"
        onClick={e => e.stopPropagation()}
        style={{ '--hero-color': hero.color } as React.CSSProperties}
      >
        <div className="hm-hero-glow" />
        {!rolling && <button className="hm-close" onClick={onClose} aria-label="Close">✕</button>}

        {/* Left — portrait (always visible, pulses while rolling) */}
        <div className={`hm-left ${rolling ? 'hm-left-rolling' : ''}`}>
          <img src={`/heroes/${hero.img}`} alt={hero.name} className="hm-portrait" />
          <div className="hm-hero-name">{hero.name}</div>
          <div className="hm-drink">
            <span className="hm-drink-icon">🍺</span>
            <div>
              <div className="hm-drink-label">Dedicate a drink when</div>
              <div className="hm-drink-rule">{hero.drinkModifier}</div>
            </div>
          </div>
        </div>

        {/* Right — same layout always, values flash while rolling */}
        <div className="hm-right">
          {/* Stats */}
          <div className="hm-stats">
            <div className="hm-stat">
              <span className="hm-stat-label">Role</span>
              <span className={`hm-stat-value ${rolling ? 'hm-flash' : 'hm-lock-in'}`}>
                {display.role}
              </span>
            </div>
            <div className="hm-stat-sep" />
            <div className="hm-stat">
              <span className="hm-stat-label">Max Ability</span>
              <span className={`hm-stat-value hm-stat-big ${rolling ? 'hm-flash' : 'hm-lock-in'}`}>
                {display.abilityMax}
              </span>
            </div>
          </div>

          {/* Can't Buy */}
          <div className="hm-section">
            <div className="hm-section-header hm-cant-header">
              <span className="hm-section-line" /><span className="hm-section-title">Can't Buy</span><span className="hm-section-line" />
            </div>
            {CATS.map(cat => (
              <div key={cat} className="hm-item-row">
                <span className="hm-item-dot" style={{ background: CAT_COLORS[cat] }} />
                <span className="hm-item-cat">{cat.charAt(0).toUpperCase() + cat.slice(1)}</span>
                <span className={`hm-item-name hm-item-cant ${rolling ? 'hm-flash' : 'hm-lock-in'}`}>
                  {display.cant[cat]}
                </span>
              </div>
            ))}
          </div>

          {/* Must Buy */}
          <div className="hm-section">
            <div className="hm-section-header hm-must-header">
              <span className="hm-section-line" /><span className="hm-section-title">Must Buy</span><span className="hm-section-line" />
            </div>
            {CATS.map(cat => (
              <div key={cat} className="hm-item-row">
                <span className="hm-item-dot" style={{ background: CAT_COLORS[cat] }} />
                <span className="hm-item-cat">{cat.charAt(0).toUpperCase() + cat.slice(1)}</span>
                <span className={`hm-item-name hm-item-must ${rolling ? 'hm-flash' : 'hm-lock-in'}`}>
                  {display.must[cat]}
                </span>
              </div>
            ))}
          </div>

          {/* Spinner bar OR reroll button */}
          {rolling ? (
            <div className="hm-rolling-bar">
              <span className="hm-roll-spinner" />
              <span className="hm-roll-spinner-label">Rolling…</span>
            </div>
          ) : (
            <button className="hm-reroll-btn" onClick={reroll}>↺&nbsp; Reroll</button>
          )}
        </div>
      </div>
    </div>
  )
}

// ── HERO ROLL ─────────────────────────────────────────────────────────────────

function HeroRoll() {
  const [activeHero, setActiveHero] = useState<Hero | null>(null)
  // key forces modal to remount (fresh roll) even when same hero is reselected
  const [modalKey, setModalKey] = useState(0)

  function open(hero: Hero) {
    setActiveHero(hero)
    setModalKey(k => k + 1)
  }

  function openRandom() {
    open(pickRandom(HEROES))
  }

  return (
    <div className="dl-roll-panel">
      <div className="dl-hero-grid">
        {HEROES.map(hero => (
          <button
            key={hero.name}
            className="dl-hero-card"
            style={{ '--hero-color': hero.color } as React.CSSProperties}
            onClick={() => open(hero)}
            title={hero.name}
          >
            <img src={`/heroes/${hero.img}`} alt={hero.name} className="dl-hero-card-img" draggable={false} />
            <span className="dl-hero-name">{hero.name}</span>
          </button>
        ))}

        {/* Random hero card */}
        <button
          className="dl-hero-card dl-hero-card-random"
          onClick={openRandom}
          title="Random Hero"
        >
          <div className="dl-random-face">?</div>
          <span className="dl-hero-name">Random</span>
        </button>
      </div>

      {activeHero && (
        <HeroModal key={modalKey} hero={activeHero} onClose={() => setActiveHero(null)} />
      )}
    </div>
  )
}

// ── PAGE ─────────────────────────────────────────────────────────────────────

export default function Deadlock() {
  const [tab, setTab] = useState<'game' | 'hero'>('game')

  return (
    <div className="dl-page">
      <div className="dl-bg-orb dl-orb-1" />
      <div className="dl-bg-orb dl-orb-2" />

      <div className="dl-header">
        <a href="/" className="dl-back">← DialedIn</a>
        <h1 className="dl-title">Deadlock Strat Roulette</h1>
      </div>

      <div className="dl-tabs">
        <button className={`dl-tab ${tab === 'game' ? 'dl-tab-active' : ''}`} onClick={() => setTab('game')}>
          🎲 Game Roll
        </button>
        <button className={`dl-tab ${tab === 'hero' ? 'dl-tab-active' : ''}`} onClick={() => setTab('hero')}>
          🦸 Hero Roll
        </button>
      </div>

      <div className="dl-content">
        {tab === 'game' ? <GameRoll /> : <HeroRoll />}
      </div>
    </div>
  )
}
