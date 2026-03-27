import './App.css'

function App() {
  return (
    <div className="landing">

      {/* ── NAV ── */}
      <nav className="nav">
        <div className="nav-container">
          <div className="nav-logo">
            <div className="logo-icon">
              <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M6 12h20l-2 14H8L6 12Z" fill="currentColor" opacity="0.15"/>
                <path d="M6 12h20l-2 14H8L6 12Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round"/>
                <path d="M22 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                <path d="M16 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                <path d="M10 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
              </svg>
            </div>
            <span className="logo-wordmark">DialedIn</span>
          </div>
          <div className="nav-links">
            <a href="#features">Features</a>
            <a href="#how-it-works">How it works</a>
            <a href="#download" className="nav-cta">Get the App</a>
          </div>
        </div>
      </nav>

      {/* ── HERO ── */}
      <section className="hero">
        <div className="hero-bg-orb hero-orb-1" />
        <div className="hero-bg-orb hero-orb-2" />
        <div className="hero-inner">
          <div className="hero-content">
            <div className="hero-badge">
              <span className="hero-badge-dot" />
              Coffee, perfected
            </div>
            <h1 className="hero-title">
              Your perfect cup,<br />
              <em>every single time.</em>
            </h1>
            <p className="hero-subtitle">
              Build precise brewing recipes, track every bean in your collection,
              and dial in your grinder — all in one beautifully simple app made
              for coffee obsessives.
            </p>
            <div className="hero-actions">
              <a href="#download" className="btn-primary">
                <svg viewBox="0 0 20 20" fill="currentColor" width="18" height="18">
                  <path d="M10 2a8 8 0 100 16A8 8 0 0010 2zm0 14a6 6 0 110-12 6 6 0 010 12z" opacity="0.3"/>
                  <path fillRule="evenodd" d="M10 4a.75.75 0 01.75.75v4.59l2.47 2.47a.75.75 0 11-1.06 1.06l-2.75-2.75A.75.75 0 019.25 9.5v-4.75A.75.75 0 0110 4z" clipRule="evenodd"/>
                </svg>
                Download on iOS
              </a>
              <a href="#features" className="btn-secondary">See features ↓</a>
            </div>
          </div>

          <div className="hero-visual">
            {/* Phone Mockup */}
            <div className="phone-mockup">
              <div className="phone-notch" />
              <div className="phone-screen">
                <div className="app-header">
                  <span className="app-back">‹</span>
                  <span className="app-title">Yirgacheffe Bloom</span>
                  <span className="app-more">•••</span>
                </div>

                <div className="timer-wrap">
                  <div className="timer-ring">
                    <svg viewBox="0 0 120 120" className="timer-svg">
                      <circle cx="60" cy="60" r="50" className="timer-track" />
                      <circle cx="60" cy="60" r="50" className="timer-progress" />
                    </svg>
                    <div className="timer-inner">
                      <span className="timer-time">2:30</span>
                      <span className="timer-phase">Bloom</span>
                    </div>
                  </div>
                  <button className="timer-btn">Pause</button>
                </div>

                <div className="app-steps">
                  <div className="app-step app-step-done">
                    <div className="app-dot app-dot-done" />
                    <div className="app-step-body">
                      <span className="app-step-name">Pre-heat</span>
                      <span className="app-step-meta">45g · 0:30</span>
                    </div>
                    <span className="app-check">✓</span>
                  </div>
                  <div className="app-step app-step-active">
                    <div className="app-dot app-dot-active" />
                    <div className="app-step-body">
                      <span className="app-step-name">Bloom</span>
                      <span className="app-step-meta">60g · 2:30</span>
                    </div>
                  </div>
                  <div className="app-step">
                    <div className="app-dot" />
                    <div className="app-step-body">
                      <span className="app-step-name">First Pour</span>
                      <span className="app-step-meta">135g · 3:00</span>
                    </div>
                  </div>
                  <div className="app-step">
                    <div className="app-dot" />
                    <div className="app-step-body">
                      <span className="app-step-name">Final Pour</span>
                      <span className="app-step-meta">80g · 4:30</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Floating cards */}
            <div className="float-card float-card-top">
              <div className="float-card-icon">⭐</div>
              <div className="float-card-body">
                <div className="float-card-title">Recipe saved!</div>
                <div className="float-card-sub">V60 · Light Roast</div>
              </div>
            </div>
            <div className="float-card float-card-bot">
              <div className="float-card-icon">⚙️</div>
              <div className="float-card-body">
                <div className="float-card-title">Grinder dialed</div>
                <div className="float-card-sub">Setting 18 · Fine</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── STATS BAR ── */}
      <div className="stats-bar">
        <div className="stat">
          <span className="stat-num">∞</span>
          <span className="stat-label">Brew Recipes</span>
        </div>
        <div className="stat-sep" />
        <div className="stat">
          <span className="stat-num">☕</span>
          <span className="stat-label">Coffee Library</span>
        </div>
        <div className="stat-sep" />
        <div className="stat">
          <span className="stat-num">⚙️</span>
          <span className="stat-label">Grinder Profiles</span>
        </div>
        <div className="stat-sep" />
        <div className="stat">
          <span className="stat-num">★</span>
          <span className="stat-label">Rate & Refine</span>
        </div>
      </div>

      {/* ── FEATURES ── */}
      <section className="features" id="features">
        <div className="section-wrap">
          <div className="section-header">
            <span className="section-badge">Features</span>
            <h2 className="section-title">Everything a coffee nerd needs</h2>
            <p className="section-sub">Built for the obsessive, refined for the everyday drinker.</p>
          </div>

          <div className="features-grid">
            <div className="feature-card feature-card-hero">
              <div className="feature-icon-wrap">
                <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28">
                  <rect x="8" y="12" width="32" height="28" rx="4" stroke="currentColor" strokeWidth="2.5"/>
                  <path d="M16 22h16M16 30h10" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                  <path d="M16 8v8M24 6v8M32 8v8" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                </svg>
              </div>
              <h3>Recipe Builder</h3>
              <p>Craft precise, step-by-step brewing recipes with timers. Bloom, pour, and rest — every detail captured exactly how you like it.</p>
              <div className="feature-chips">
                <span className="chip">Step timers</span>
                <span className="chip">Water ratios</span>
                <span className="chip">Temperature</span>
                <span className="chip">Guided execution</span>
              </div>
            </div>

            <div className="feature-card">
              <div className="feature-icon-wrap">
                <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28">
                  <path d="M10 14h20v22a4 4 0 01-4 4H14a4 4 0 01-4-4V14z" stroke="currentColor" strokeWidth="2.5"/>
                  <path d="M30 18h4a4 4 0 010 8h-4" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                  <path d="M10 8h20" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                </svg>
              </div>
              <h3>Coffee Library</h3>
              <p>Track every bean you've tried. Log origins, roasters, roast levels, and tasting notes — your personal encyclopedia of coffees.</p>
              <div className="feature-chips">
                <span className="chip">Origins</span>
                <span className="chip">Tasting notes</span>
                <span className="chip">Ratings</span>
              </div>
            </div>

            <div className="feature-card">
              <div className="feature-icon-wrap">
                <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" width="28" height="28">
                  <circle cx="24" cy="24" r="14" stroke="currentColor" strokeWidth="2.5"/>
                  <circle cx="24" cy="24" r="5" stroke="currentColor" strokeWidth="2.5"/>
                  <path d="M24 6v6M24 36v6M6 24h6M36 24h6" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                </svg>
              </div>
              <h3>Grinder Profiles</h3>
              <p>Save grinder settings for every coffee and brew method. Never forget the setting that gave you that perfect extraction.</p>
              <div className="feature-chips">
                <span className="chip">Settings</span>
                <span className="chip">Burr type</span>
                <span className="chip">Per-coffee</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── HOW IT WORKS ── */}
      <section className="how-it-works" id="how-it-works">
        <div className="section-wrap">
          <div className="section-header">
            <span className="section-badge">How it works</span>
            <h2 className="section-title">From bean to cup, in four steps</h2>
            <p className="section-sub">An intentional flow designed around how great coffee actually gets made.</p>
          </div>

          <div className="timeline">
            <div className="timeline-step">
              <div className="timeline-node">
                <div className="timeline-num">01</div>
                <div className="timeline-connector" />
              </div>
              <div className="timeline-card">
                <div className="timeline-icon">☕</div>
                <h3>Add your coffee</h3>
                <p>Log your beans with origin, roaster, roast date, and tasting notes. Build a library that grows with every bag you open.</p>
              </div>
            </div>

            <div className="timeline-step">
              <div className="timeline-node">
                <div className="timeline-num">02</div>
                <div className="timeline-connector" />
              </div>
              <div className="timeline-card">
                <div className="timeline-icon">⚙️</div>
                <h3>Dial in your grinder</h3>
                <p>Save your grinder settings per coffee and brew method. Hit that sweet spot once, and you can always go back to it.</p>
              </div>
            </div>

            <div className="timeline-step">
              <div className="timeline-node">
                <div className="timeline-num">03</div>
                <div className="timeline-connector" />
              </div>
              <div className="timeline-card">
                <div className="timeline-icon">📋</div>
                <h3>Build a recipe</h3>
                <p>Design step-by-step brewing recipes with precise timings, water amounts, and temperatures — as simple or detailed as you like.</p>
              </div>
            </div>

            <div className="timeline-step">
              <div className="timeline-node">
                <div className="timeline-num">04</div>
              </div>
              <div className="timeline-card">
                <div className="timeline-icon">★</div>
                <h3>Brew, rate & refine</h3>
                <p>Execute your recipe with built-in guided timers. Rate the result and tweak until every cup is exactly what you want.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── RECIPE HIGHLIGHT ── */}
      <section className="highlight">
        <div className="section-wrap">
          <div className="highlight-grid">
            <div className="highlight-visual">
              <div className="recipe-demo">
                <div className="rd-top">
                  <div className="rd-info">
                    <span className="rd-name">Ethiopian Yirgacheffe</span>
                    <span className="rd-method">V60 Pour Over · Light Roast</span>
                  </div>
                  <div className="rd-stars">★★★★★</div>
                </div>

                <div className="rd-params">
                  <div className="rd-param">
                    <span className="rd-val">20g</span>
                    <span className="rd-key">Coffee</span>
                  </div>
                  <div className="rd-param">
                    <span className="rd-val">320g</span>
                    <span className="rd-key">Water</span>
                  </div>
                  <div className="rd-param">
                    <span className="rd-val">93°C</span>
                    <span className="rd-key">Temp</span>
                  </div>
                  <div className="rd-param">
                    <span className="rd-val">4:30</span>
                    <span className="rd-key">Time</span>
                  </div>
                </div>

                <div className="rd-divider" />

                <div className="rd-steps">
                  <div className="rd-step rd-step-done">
                    <div className="rd-dot rd-dot-done" />
                    <div className="rd-step-line" />
                    <span className="rd-step-text">Bloom — 45g · 0:30</span>
                    <span className="rd-check">✓</span>
                  </div>
                  <div className="rd-step rd-step-active">
                    <div className="rd-dot rd-dot-active" />
                    <div className="rd-step-line" />
                    <span className="rd-step-text">First pour — 135g · 2:00</span>
                    <span className="rd-badge">Active</span>
                  </div>
                  <div className="rd-step">
                    <div className="rd-dot" />
                    <div className="rd-step-line" />
                    <span className="rd-step-text">Final pour — 140g · 4:30</span>
                  </div>
                </div>

                <button className="rd-execute-btn">Execute Recipe</button>
              </div>
            </div>

            <div className="highlight-text">
              <span className="section-badge">Recipe execution</span>
              <h2>Your recipes, brought to life</h2>
              <p>
                Step through your brewing recipe in real time with built-in timers
                for each phase. Bloom, pour, and rest — guided precisely so you
                can focus on the craft.
              </p>
              <ul className="check-list">
                <li>
                  <span className="check-icon">✓</span>
                  Step-by-step guided timers
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Water weight tracking per pour
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Temperature reminders
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Post-brew ratings & notes
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Save & share your recipes
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* ── GRINDER HIGHLIGHT ── */}
      <section className="highlight highlight-alt">
        <div className="section-wrap">
          <div className="highlight-grid highlight-grid-rev">
            <div className="highlight-text">
              <span className="section-badge">Grinder management</span>
              <h2>Never lose your perfect setting</h2>
              <p>
                Track grinder settings alongside each coffee so you can
                consistently reproduce your best brews. Different bean?
                Different method? Save a separate profile for each.
              </p>
              <ul className="check-list">
                <li>
                  <span className="check-icon">✓</span>
                  Per-coffee grinder settings
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Multiple grinder support
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Burr type & grind notes
                </li>
                <li>
                  <span className="check-icon">✓</span>
                  Quick-access profiles
                </li>
              </ul>
            </div>

            <div className="highlight-visual">
              <div className="grinder-demo">
                <div className="gd-header">
                  <span className="gd-title">1Zpresso J-Max</span>
                  <span className="gd-badge">Active</span>
                </div>
                <div className="gd-dial-wrap">
                  <div className="gd-dial">
                    <div className="gd-dial-ring" />
                    <div className="gd-dial-inner">
                      <span className="gd-setting">18</span>
                      <span className="gd-unit">clicks</span>
                    </div>
                    <div className="gd-dial-marker" />
                  </div>
                </div>
                <div className="gd-tags">
                  <span className="gd-tag">Fine</span>
                  <span className="gd-tag">Espresso</span>
                  <span className="gd-tag">Ethiopian</span>
                </div>
                <div className="gd-notes">
                  <span className="gd-notes-label">Notes</span>
                  <span className="gd-notes-text">Sweet spot for light roast V60. Bright and clean with no bitterness.</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── CTA ── */}
      <section className="cta-section" id="download">
        <div className="cta-bg-orb" />
        <div className="section-wrap">
          <div className="cta-card">
            <div className="cta-glyph">☕</div>
            <h2>Ready to dial it in?</h2>
            <p>
              Join the coffee enthusiasts who've upgraded their brewing game.
              Free to download. No subscription required.
            </p>
            <div className="cta-actions">
              <a href="#" className="btn-primary btn-xl">
                <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
                  <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
                </svg>
                Download on iOS
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* ── FOOTER ── */}
      <footer className="footer">
        <div className="footer-inner">
          <div className="footer-logo">
            <div className="logo-icon logo-icon-sm">
              <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M6 12h20l-2 14H8L6 12Z" fill="currentColor" opacity="0.15"/>
                <path d="M6 12h20l-2 14H8L6 12Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round"/>
                <path d="M22 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                <path d="M16 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
                <path d="M10 12c0-3 3-3 3-6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
              </svg>
            </div>
            <span className="logo-wordmark">DialedIn</span>
          </div>
          <p className="footer-tagline">Your perfect cup, every time.</p>
          <p className="footer-copy">© 2025 DialedIn. Built for coffee lovers.</p>
        </div>
      </footer>

    </div>
  )
}

export default App
