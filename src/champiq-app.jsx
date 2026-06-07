/* ChampIQ — Vite/React port: kit + 14 screens concatenated into one module
   (preserves the authored shared-scope code; React provided via imports). */
import React from 'react'
import ReactDOM from 'react-dom/client'


/* ======== kit.jsx ======== */
/* ChampIQ UI Kit — shared primitives & helpers.
   Self-contained recreation of the product's atoms so the kit renders
   standalone. Mirrors the design-system components 1:1 in styling. */

/* ── Icon set (subset used across the kit) ─────────────────────────────── */
const ICON_PATHS = {
  chat:<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>,
  mail:<><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 7l9 6 9-6"/></>,
  graph:<><circle cx="6" cy="6" r="2.5"/><circle cx="18" cy="6" r="2.5"/><circle cx="12" cy="18" r="2.5"/><path d="M7.6 7.4l3 8.4M16.4 7.4l-3 8.4M8 6h8"/></>,
  settings:<><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9 1.7 1.7 0 0 0 4.3 7.2l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/></>,
  play:<polygon points="5 3 19 12 5 21 5 3"/>,
  save:<><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></>,
  chevRight:<polyline points="9 18 15 12 9 6"/>,
  chevLeft:<polyline points="15 18 9 12 15 6"/>,
  plus:<><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></>,
  x:<><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></>,
  sparkle:<path d="M12 2l2 6 6 2-6 2-2 6-2-6-6-2 6-2z"/>,
  bolt:<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>,
  search:<><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></>,
  send:<><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></>,
  user:<><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></>,
  folder:<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>,
  layers:<><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></>,
  loop:<><path d="M3 12a9 9 0 1 0 3-6.7"/><polyline points="3 4 3 9 8 9"/></>,
  if_node:<><path d="M12 3v6"/><path d="M12 9 7 16h10z"/><path d="M7 16v3M17 16v3"/></>,
  webhook:<><path d="M9 9a3 3 0 1 1 4 2.7L7 18"/><path d="M16 13a3 3 0 1 1-3 3h-7"/><path d="M12 6a3 3 0 1 1 3 3v0"/></>,
  voice:<><rect x="9" y="3" width="6" height="12" rx="3"/><path d="M5 11a7 7 0 0 0 14 0"/><line x1="12" y1="18" x2="12" y2="22"/></>,
  play_node:<><circle cx="12" cy="12" r="9"/><polygon points="10 8 16 12 10 16 10 8" fill="currentColor"/></>,
  branch:<><path d="M6 3v18M18 8a4 4 0 0 0-4-4H6"/><circle cx="6" cy="3" r="1.5" fill="currentColor"/><circle cx="18" cy="8" r="1.5" fill="currentColor"/><circle cx="6" cy="21" r="1.5" fill="currentColor"/></>,
  set:<><rect x="3" y="4" width="18" height="4" rx="1"/><rect x="3" y="10" width="14" height="4" rx="1"/><rect x="3" y="16" width="10" height="4" rx="1"/></>,
  db:<><ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v6c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 11v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6"/></>,
  users:<><circle cx="9" cy="8" r="3.2"/><path d="M3 20a6 6 0 0 1 12 0"/><path d="M16 5.5a3.2 3.2 0 0 1 0 6M21 20a6 6 0 0 0-4-5.6"/></>,
  clock:<><circle cx="12" cy="12" r="9"/><polyline points="12 7 12 12 15 14"/></>,
  check:<polyline points="20 6 9 17 4 12"/>,
  archive:<><path d="M3 6h18"/><rect x="3" y="6" width="18" height="14" rx="1"/><line x1="10" y1="12" x2="14" y2="12"/></>,
  grid:<><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></>,
  home:<><path d="M3 12 12 3l9 9"/><path d="M5 10v10h14V10"/></>,
}
function Icon({ name, size = 18, stroke = 'currentColor', strokeWidth = 1.6, style }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={stroke}
      strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round" style={style}>
      {ICON_PATHS[name] ?? null}
    </svg>
  )
}

/* ── Button ─────────────────────────────────────────────────────────────── */
const BTN_VAR = {
  primary:{ background:'linear-gradient(180deg, var(--accent-2), var(--accent-3))', color:'#fff', border:'1px solid var(--accent-3)', boxShadow:'0 0 24px -6px rgba(var(--accent-2-rgb),.6), inset 0 1px 0 rgba(255,255,255,.18)' },
  pixie:{ background:'var(--accent-2)', color:'#fff', border:'1px solid var(--accent-3)', boxShadow:'0 0 24px -8px rgba(var(--accent-2-rgb),.6)' },
  secondary:{ background:'var(--bg-2)', color:'var(--text-1)', border:'1px solid var(--border-1)' },
  ghost:{ background:'transparent', color:'var(--text-2)', border:'1px solid transparent' },
  danger:{ background:'rgba(255,77,109,.14)', color:'var(--danger)', border:'1px solid rgba(255,77,109,.4)' },
}
const BTN_SIZE = {
  sm:{ padding:'4px 10px', fontSize:12, borderRadius:6, height:26 },
  md:{ padding:'6px 14px', fontSize:13, borderRadius:8, height:32 },
  lg:{ padding:'10px 18px', fontSize:14, borderRadius:10, height:40 },
}
function Button({ variant='secondary', size='md', icon, kbd, children, style, ...rest }) {
  return (
    <button style={{ display:'inline-flex', alignItems:'center', gap:8, fontFamily:'var(--font-display)',
      fontWeight:600, letterSpacing:'.02em', cursor:rest.disabled?'not-allowed':'pointer', whiteSpace:'nowrap',
      opacity:rest.disabled?0.5:1, ...BTN_VAR[variant], ...BTN_SIZE[size], ...style }} {...rest}>
      {icon && <Icon name={icon} size={size==='sm'?12:14} />}
      {children}
      {kbd && <span style={{ fontFamily:'var(--font-mono)', fontSize:10, opacity:.65, marginLeft:4, background:'rgba(0,0,0,.25)', padding:'1px 5px', borderRadius:3 }}>{kbd}</span>}
    </button>
  )
}

/* ── Tag & Badge ────────────────────────────────────────────────────────── */
function Tag({ children, color='var(--accent-2)', filled=false, style }) {
  return <span style={{ display:'inline-flex', alignItems:'center', gap:4, padding:'2px 7px', borderRadius:4,
    fontFamily:'var(--font-mono)', fontSize:10, letterSpacing:'.12em', textTransform:'uppercase', fontWeight:600,
    whiteSpace:'nowrap', background:filled?color:`color-mix(in oklch, ${color} 16%, transparent)`, color:filled?'var(--bg-0)':color,
    border:filled?'none':`1px solid color-mix(in oklch, ${color} 35%, transparent)`, ...style }}>{children}</span>
}
const BADGE_STATUS = {
  idle:{ color:'var(--text-3)', label:'Idle' }, running:{ color:'var(--warn)', label:'Running' },
  success:{ color:'var(--success)', label:'Done' }, error:{ color:'var(--danger)', label:'Errors' },
  paused:{ color:'var(--text-4)', label:'Paused' }, draft:{ color:'var(--accent-2)', label:'Draft' },
}
function Badge({ status='idle', color, label, pulse=false, style }) {
  const s = BADGE_STATUS[status] ?? BADGE_STATUS.idle, c = color ?? s.color, t = label ?? s.label
  return <span style={{ display:'inline-flex', alignItems:'center', gap:5, padding:'2px 8px', borderRadius:999,
    background:`color-mix(in oklch, ${c} 14%, transparent)`, border:`1px solid color-mix(in oklch, ${c} 30%, transparent)`,
    fontFamily:'var(--font-mono)', fontSize:9, textTransform:'uppercase', letterSpacing:'.1em', color:c, ...style }}>
    <span style={{ width:5, height:5, borderRadius:'50%', background:c, boxShadow:`0 0 6px ${c}`, animation:pulse?'glow-pulse 1s ease-in-out infinite':'none' }} />{t}</span>
}

/* ── Input ──────────────────────────────────────────────────────────────── */
function Input({ icon, style, onFocus, onBlur, ...rest }) {
  const [f, setF] = React.useState(false)
  return (
    <div style={{ display:'inline-flex', alignItems:'center', gap:8, width:'100%', background:'var(--bg-2)',
      border:`1px solid ${f?'var(--accent-2)':'var(--border-1)'}`, borderRadius:8, padding:'7px 12px',
      boxShadow:f?'0 0 0 3px rgba(var(--accent-2-rgb),.16)':'none', transition:'border-color .15s, box-shadow .15s', ...style }}>
      {icon && <Icon name={icon} size={14} stroke="var(--text-3)" />}
      <input {...rest} onFocus={e=>{setF(true);onFocus?.(e)}} onBlur={e=>{setF(false);onBlur?.(e)}}
        style={{ flex:1, minWidth:0, background:'transparent', border:'none', outline:'none', color:'var(--text-1)', fontFamily:'var(--font-body)', fontSize:13 }} />
    </div>
  )
}

/* ── Brand: ChampMark + Wordmark ────────────────────────────────────────── */
function ChampMark({ size=22, style }) {
  return (
    <svg viewBox="0 0 16 16" width={size} height={size} shapeRendering="crispEdges" style={style}>
      <rect x="2" y="6" width="3" height="3" fill="var(--accent-2)"/><rect x="5" y="3" width="3" height="3" fill="var(--accent-1)"/>
      <rect x="8" y="6" width="3" height="3" fill="var(--accent-2)"/><rect x="5" y="9" width="3" height="3" fill="var(--accent-3)"/>
      <rect x="11" y="9" width="3" height="3" fill="var(--mint-2)"/>
    </svg>
  )
}
function Wordmark({ size=22, mark=true, style }) {
  return <div style={{ display:'flex', alignItems:'center', gap:8, ...style }}>
    {mark && <ChampMark size={size} />}
    <span style={{ fontFamily:'var(--font-display)', fontSize:size, fontWeight:700, letterSpacing:'-.025em', color:'var(--text-1)', lineHeight:1 }}>Champ<span style={{ color:'var(--accent-2)' }}>IQ</span></span>
  </div>
}

/* ── Card ───────────────────────────────────────────────────────────────── */
function Card({ children, hover=false, accentBar, padding=16, style, ...rest }) {
  const [over,setOver] = React.useState(false)
  return (
    <div onMouseEnter={()=>setOver(true)} onMouseLeave={()=>setOver(false)} style={{ position:'relative', background:'var(--bg-1)', border:'1px solid var(--border-1)', borderColor:hover&&over?(accentBar??'var(--accent-2)'):'var(--border-1)', borderRadius:12, padding, transition:'transform .18s var(--ease-swift), border-color .18s, box-shadow .18s', transform:hover&&over?'translateY(-2px)':'none', boxShadow:hover&&over?`0 8px 28px -10px ${accentBar??'var(--accent-2)'}88`:'none', overflow:accentBar?'hidden':'visible', ...style }} {...rest}>
      {accentBar && <div style={{ position:'absolute', top:0, left:0, right:0, height:3, background:accentBar, borderRadius:'12px 12px 0 0' }} />}
      {children}
    </div>
  )
}

/* ── Switch ─────────────────────────────────────────────────────────────── */
function Switch({ checked=false, onChange, disabled=false, size='md', style }) {
  const w = size==='sm'?30:38, h = size==='sm'?18:22, knob = h-6
  return (
    <button role="switch" aria-checked={checked} disabled={disabled} onClick={()=>!disabled&&onChange?.(!checked)} style={{ width:w, height:h, borderRadius:999, padding:0, position:'relative', border:'1px solid', cursor:disabled?'not-allowed':'pointer', background:checked?'var(--accent-2)':'var(--bg-3)', borderColor:checked?'var(--accent-3)':'var(--border-2)', opacity:disabled?0.5:1, boxShadow:checked?'0 0 14px -4px rgba(var(--accent-2-rgb),.7)':'none', transition:'background .18s var(--ease-swift), border-color .18s, box-shadow .18s', ...style }}>
      <span style={{ position:'absolute', top:2, left:checked?w-knob-3:2, width:knob, height:knob, borderRadius:'50%', background:'#fff', boxShadow:'0 1px 3px rgba(0,0,0,.4)', transition:'left .18s var(--ease-spring)' }} />
    </button>
  )
}

/* ── Pixie ──────────────────────────────────────────────────────────────── */
function Pixie({ pose='idle', cloak='cobalt', size=96, ambient=true, flip=false, basePath='//assets/pixie', style }) {
  const src = pose==='idle' ? `${basePath}/pixie-${cloak}.png` : `${basePath}/pixie-${cloak}-${pose}.png`
  const outer = !ambient || pose==='cheer' ? 'none' : pose==='sleep' ? 'pixie-float 6s ease-in-out infinite' : 'pixie-float 4s ease-in-out infinite'
  return (
    <div aria-hidden="true" style={{ position:'relative', display:'inline-flex', flexDirection:'column', alignItems:'center', animation:outer, ...style }}>
      <div style={{ position:'relative', width:size, height:size, animation:ambient?'pixie-breathe 3.2s ease-in-out infinite':'none', transformOrigin:'50% 85%', transform:flip?'scaleX(-1)':'none' }}>
        <div style={{ position:'absolute', left:'50%', bottom:-2, transform:'translateX(-50%)', width:size*0.6, height:size*0.08, borderRadius:'50%', background:'rgba(30,95,203,.28)', filter:'blur(4px)', animation:ambient?'glow-pulse 3.2s ease-in-out infinite':'none' }} />
        <img src={src} alt="" className="pixie-sprite" style={{ width:size, height:size, objectFit:'contain', position:'relative' }} />
        {ambient && pose!=='sleep' && <div style={{ position:'absolute', left:'38%', top:'8%', width:'24%', height:'14%', background:'radial-gradient(ellipse at center, rgba(255,210,63,.55), transparent 65%)', animation:'antenna-spark 1.6s ease-in-out infinite', pointerEvents:'none' }} />}
      </div>
    </div>
  )
}

/* ── NodeCard ───────────────────────────────────────────────────────────── */
const KIND_COLOR = { trigger:'#10b981', cron:'#10b981', webhook:'#10b981', set:'#06b6d4', merge:'#06b6d4', csv:'#06b6d4', if:'#f59e0b', switch:'#f59e0b', loop:'#f59e0b', split:'#f59e0b', classifier:'#f59e0b', wait:'#6b7280', code:'#6b7280', http:'#8b5cf6', llm:'#8b5cf6', champmail:'#f97316', champgraph:'#14b8a6', champvoice:'#3b82f6', harbinger:'#E63A87', approval:'#5BC0FF' }
const KIND_ICON = { trigger:'play_node', cron:'clock', webhook:'webhook', set:'set', merge:'layers', csv:'db', if:'if_node', switch:'branch', loop:'loop', split:'branch', classifier:'branch', wait:'clock', code:'set', http:'webhook', llm:'sparkle', champmail:'mail', champgraph:'graph', champvoice:'voice', harbinger:'search', approval:'user' }
const ST_BORDER = { idle:'3px solid var(--border-2)', running:'3px solid var(--warn)', success:'3px solid var(--success)', error:'3px solid var(--danger)', waiting:'3px dashed var(--info)', suggesting:'3px dashed var(--mint-2)' }
const ST_TEXT = { running:'Running…', success:'Done', error:'Error', waiting:'Waiting…', suggesting:'Pixie suggests' }
const ST_COLOR = { running:'var(--warn)', success:'var(--success)', error:'var(--danger)', waiting:'var(--info)', suggesting:'var(--mint-2)' }
function NodeHandle({ side }) {
  return <div style={{ position:'absolute', top:'50%', [side]:-6, transform:'translateY(-50%)', width:12, height:12, borderRadius:'50%', background:'var(--bg-3)', border:'2px solid var(--border-2)' }} />
}
function NodeCard({ kind='llm', kindLabel, label='Untitled node', status='idle', selected=false, showTarget=true, showSource=true, style, ...rest }) {
  const color = KIND_COLOR[kind] ?? 'var(--accent-2)', icon = KIND_ICON[kind] ?? 'layers'
  const isTrigger = kind==='trigger'||kind==='cron'||kind==='webhook'
  return (
    <div style={{ width:200, minHeight:84, position:'relative', background:'var(--bg-1)', border:'1px solid var(--border-2)', borderLeft:ST_BORDER[status]??ST_BORDER.idle, borderRadius:12, padding:'10px 12px', cursor:'pointer', boxShadow:selected?'0 0 0 2px var(--accent-2), 0 0 24px -4px rgba(var(--accent-2-rgb),.4)':'none', transition:'box-shadow .18s var(--ease-swift)', ...style }} {...rest}>
      {showTarget && !isTrigger && <NodeHandle side="left" />}
      <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:status!=='idle'?6:0 }}>
        <div style={{ width:28, height:28, borderRadius:7, flexShrink:0, background:`color-mix(in oklch, ${color} 22%, transparent)`, color, display:'grid', placeItems:'center' }}><Icon name={icon} size={14} /></div>
        <div style={{ flex:1, minWidth:0 }}>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)' }}>{kindLabel??kind}</div>
          <div style={{ fontFamily:'var(--font-display)', fontSize:12.5, fontWeight:600, color:'var(--text-1)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{label}</div>
        </div>
      </div>
      {status!=='idle' && <div style={{ fontFamily:'var(--font-mono)', fontSize:10, letterSpacing:'.08em', color:ST_COLOR[status]??'var(--text-3)' }}>{ST_TEXT[status]??status}</div>}
      {showSource && <NodeHandle side="right" />}
    </div>
  )
}

Object.assign(window, { Icon, Button, Tag, Badge, Input, Card, Switch, ChampMark, Wordmark, Pixie, NodeCard, KIND_COLOR })

/* ======== data.jsx ======== */
/* ChampIQ (Sim) — mock data: clients, knowledge graphs, canvases, signals. */

const CLIENTS = [
  { id:'acme', name:'Acme SaaS', icp:'B2B SaaS · 50–200 staff · hiring VP Sales', accent:'#7C5CFF',
    health:78, prospects:412, signals:38, replies:11 },
  { id:'nimbus', name:'Nimbus Fintech', icp:'Fintech ops leaders · Series B+', accent:'#00E5C7',
    health:64, prospects:286, signals:21, replies:7 },
  { id:'vela', name:'Vela Health', icp:'Healthtech · RevOps · 200+ staff', accent:'#5BC0FF',
    health:71, prospects:198, signals:14, replies:5 },
]

// Per-client knowledge graph: prospects with fit/intent + signals + stakeholder edges.
const GRAPHS = {
  acme: {
    companies: [
      { id:'northwind', name:'Northwind', employees:140, health:86, signals:['hiring_sdrs','pricing_visit'] },
      { id:'oakline',   name:'Oakline',   employees:90,  health:72, signals:['funding_round'] },
      { id:'deadlink',  name:'Deadlink',  employees:60,  health:22, signals:[] },
    ],
    prospects: [
      { id:'dana', name:'Dana Okafor',  title:'VP Sales',      company:'northwind', fit:100, intent:'hot',  signals:'hiring_sdrs · pricing-visit', reportsTo:null },
      { id:'sam',  name:'Sam Cole',     title:'Head of Growth', company:'oakline',  fit:82,  intent:'hot',  signals:'funding_round', reportsTo:null },
      { id:'mia',  name:'Mia Chen',     title:'SDR Manager',    company:'northwind', fit:74,  intent:'warm', signals:'team-scaling', reportsTo:'dana' },
      { id:'pat',  name:'Pat Reyes',    title:'Ops',            company:'deadlink',  fit:40,  intent:'cold', signals:'— (suppressed)', reportsTo:null },
    ],
  },
  nimbus: {
    companies: [
      { id:'vaultpay', name:'VaultPay', employees:320, health:81, signals:['compliance_hiring','soc2'] },
    ],
    prospects: [
      { id:'lee', name:'Lee Park', title:'COO', company:'vaultpay', fit:88, intent:'hot', signals:'compliance_hiring · SOC2', reportsTo:null },
      { id:'jo',  name:'Jo Mensah', title:'Head of Risk', company:'vaultpay', fit:69, intent:'warm', signals:'audit-cycle', reportsTo:'lee' },
    ],
  },
  vela: {
    companies: [ { id:'curaspan', name:'Curaspan', employees:240, health:70, signals:['rev_ops_hire'] } ],
    prospects: [ { id:'ira', name:'Ira Solis', title:'VP RevOps', company:'curaspan', fit:79, intent:'hot', signals:'rev_ops_hire', reportsTo:null } ],
  },
}

const INTENT_COLOR = { hot:'var(--success)', warm:'var(--warn)', cold:'var(--text-4)' }

// Canvas graphs (node/edge) — the signal-first SDR DAG is the headline.
const CANVAS_GRAPHS = {
  sdr: {
    nodes: [
      { id:'n1', kind:'harbinger',  label:'Discover + signals', x:40,  y:200, status:'idle' },
      { id:'n2', kind:'champgraph', label:'Enrich → client KG', x:280, y:200, status:'idle' },
      { id:'n3', kind:'llm',        label:'Signal-personalize', x:520, y:200, status:'idle' },
      { id:'n4', kind:'champmail',  label:'Send sequence',      x:760, y:200, status:'idle' },
      { id:'n5', kind:'wait',       label:'Wait for reply',     x:1000, y:200, status:'idle' },
      { id:'n6', kind:'classifier', label:'Reply classifier',   x:1240, y:200, status:'idle' },
      { id:'n7', kind:'approval',   label:'Approve call',       x:1480, y:120, status:'idle' },
      { id:'n8', kind:'champvoice', label:'ChampVoice (gated)', x:1480, y:300, status:'idle' },
      { id:'n9', kind:'champgraph', label:'Log to graph',       x:1720, y:200, status:'idle' },
    ],
    edges: [
      { from:'n1', to:'n2', state:'idle' }, { from:'n2', to:'n3', state:'idle' },
      { from:'n3', to:'n4', state:'idle' }, { from:'n4', to:'n5', state:'idle' },
      { from:'n5', to:'n6', state:'idle' }, { from:'n6', to:'n7', state:'idle' },
      { from:'n7', to:'n8', state:'idle' }, { from:'n8', to:'n9', state:'idle' },
      { from:'n6', to:'n9', state:'idle' },
    ],
  },
  classifier: {
    nodes: [
      { id:'n1', kind:'webhook',    label:'Inbox webhook',    x:60,  y:170, status:'idle' },
      { id:'n2', kind:'llm',        label:'Classify intent',  x:300, y:170, status:'idle' },
      { id:'n3', kind:'if',         label:'Positive reply?',  x:540, y:170, status:'idle' },
      { id:'n4', kind:'approval',   label:'Approve before call', x:800, y:90, status:'waiting' },
      { id:'n5', kind:'champmail',  label:'Send nudge',       x:800, y:250, status:'idle' },
    ],
    edges: [ { from:'n1', to:'n2', state:'idle' }, { from:'n2', to:'n3', state:'idle' }, { from:'n3', to:'n4', state:'idle' }, { from:'n3', to:'n5', state:'idle' } ],
  },
}

const CANVASES = [
  { id:'c1', name:'Signal-first SDR · Acme',  client:'acme',   nodes:9, updated:'2m ago', status:'running', graph:CANVAS_GRAPHS.sdr },
  { id:'c2', name:'Reply classifier',          client:'acme',   nodes:5, updated:'18m ago', status:'success', graph:CANVAS_GRAPHS.classifier },
  { id:'c3', name:'Compliance warmup · Nimbus', client:'nimbus', nodes:9, updated:'1h ago', status:'draft', graph:CANVAS_GRAPHS.sdr },
  { id:'c4', name:'RevOps plays · Vela',        client:'vela',   nodes:5, updated:'3h ago', status:'paused', graph:CANVAS_GRAPHS.classifier },
]

const TEMPLATES = [
  { title:'Signal-first SDR', desc:'Harbinger → KG enrich → personalize → mail → reply branch → voice.', tag:'Sales', color:'#F97316', graph:'sdr' },
  { title:'Reply classifier', desc:'Inbox → LLM categorize → human gate → branch.', tag:'AI', color:'var(--mint-2)', graph:'classifier' },
  { title:'Lead enrichment', desc:'CSV → ChampGraph join → score → write back.', tag:'Data', color:'#5BC0FF', graph:'sdr' },
  { title:'Voice screener', desc:'ChampVoice transcript → score → log to graph.', tag:'Voice', color:'#A855F7', graph:'classifier' },
]

// SDR run flow stages (for the run timeline / play).
const FLOW = [
  { kind:'harbinger',  t:'Discover + signals', d:'ChampHarbinger finds in-ICP companies and live buying signals (funding, exec-change, tech-install).' },
  { kind:'champgraph', t:'Enrich → client KG', d:"Pull the prospect's node from the client's isolated graph → fit score + intent band. Grounds everything downstream." },
  { kind:'llm',        t:'Signal-personalize', d:'Generate subject + body from the KG node\u2019s signals & intent — "saw Northwind is scaling the sales team", not a static template.' },
  { kind:'champmail',  t:'Send sequence',      d:'Suppression gate → per-client domain → warmup cap → own MTA. One-click unsubscribe headers injected.' },
  { kind:'wait',       t:'Wait (snapshot)',    d:'Sim serializes execution state and pauses — no held process for the day-long gap. Resumes on the inbound reply webhook.' },
  { kind:'classifier', t:'Reply classifier',   d:'Inbound reply → cheap LLM classifies positive/negative/neutral → branch. Positive pauses the sequence.' },
  { kind:'approval',   t:'Human approval',     d:'A human-in-the-loop gate precedes high-value calls (45% of top teams run hybrid).' },
  { kind:'champvoice', t:'ChampVoice (gated)', d:'Fires only for hot intent AND voice consent (DND/TCPA gate). KG-personalized ElevenLabs script.' },
  { kind:'champgraph', t:'Log to graph',       d:'Outcome writes back as edges on the prospect node — the graph compounds, improving future fit + personalization.' },
]

// zerolang checked-edit verdicts (mirrors the guardrail).
const GUARD_CASES = [
  { id:'valid',        label:'Add edge — valid hash, endpoints exist', ok:true,  code:'STALE_HASH ✓  ·  EXPECT ✓  ·  EDGE ✓  ·  GATE ✓', msg:'Accepted — hash matches, endpoints exist, capability present.' },
  { id:'stale',        label:'Add edge — STALE graphHash', ok:false, code:'reject(STALE_HASH)', msg:'Rejected [STALE_HASH] — the graph changed since this edit was authored. Re-read & retry.' },
  { id:'dangling',     label:'Add edge to a missing node', ok:false, code:'reject(DANGLING_EDGE)', msg:'Rejected [DANGLING_EDGE] — edge endpoint does not exist on the graph.' },
  { id:'voice_nogate', label:'Add ChampVoice WITHOUT consent', ok:false, code:'reject(COMPLIANCE_GATE)', msg:'Rejected [COMPLIANCE_GATE] — champvoice needs a consent.voice / DND capability declared.' },
  { id:'voice_gate',   label:'Add ChampVoice AFTER declaring consent.voice', ok:true, code:'caps.add(consent.voice) ✓', msg:'Accepted — consent.voice capability now declared on the workflow.' },
]

// Flattened prospects per client (table view) — derived + extended from GRAPHS.
const PROSPECTS = {
  acme: [
    { id:'dana', name:'Dana Okafor', title:'VP Sales', company:'Northwind', fit:100, intent:'hot', stage:'Replied', signals:['hiring_sdrs','pricing_visit'], suppressed:false, seq:'Signal-first SDR' },
    { id:'sam',  name:'Sam Cole', title:'Head of Growth', company:'Oakline', fit:82, intent:'hot', stage:'Sequenced', signals:['funding_round'], suppressed:false, seq:'Signal-first SDR' },
    { id:'mia',  name:'Mia Chen', title:'SDR Manager', company:'Northwind', fit:74, intent:'warm', stage:'Enriched', signals:['team_scaling'], suppressed:false, seq:'—' },
    { id:'leo',  name:'Leo Frank', title:'RevOps Lead', company:'Brightway', fit:66, intent:'warm', stage:'Enriched', signals:['tech_install'], suppressed:false, seq:'—' },
    { id:'pat',  name:'Pat Reyes', title:'Ops', company:'Deadlink', fit:40, intent:'cold', stage:'Suppressed', signals:[], suppressed:true, seq:'—' },
    { id:'noor', name:'Noor Ali', title:'CMO', company:'Lumen Labs', fit:58, intent:'cold', stage:'New', signals:[], suppressed:false, seq:'—' },
  ],
  nimbus: [
    { id:'lee', name:'Lee Park', title:'COO', company:'VaultPay', fit:88, intent:'hot', stage:'Sequenced', signals:['compliance_hiring','soc2'], suppressed:false, seq:'Compliance warmup' },
    { id:'jo',  name:'Jo Mensah', title:'Head of Risk', company:'VaultPay', fit:69, intent:'warm', stage:'Enriched', signals:['audit_cycle'], suppressed:false, seq:'—' },
  ],
  vela: [
    { id:'ira', name:'Ira Solis', title:'VP RevOps', company:'Curaspan', fit:79, intent:'hot', stage:'Replied', signals:['rev_ops_hire'], suppressed:false, seq:'RevOps plays' },
  ],
}

// Inbox threads — classified replies awaiting triage / approval.
const INBOX = {
  acme: [
    { id:'t1', from:'Dana Okafor', co:'Northwind', title:'VP Sales', cls:'positive', snippet:'Re: saw you\u2019re scaling the sales team — yes, happy to chat. Thursday?', body:'Hi — yes, we just opened two SDR reqs so the timing is good. Thursday afternoon works. Can you send a couple of slots?', why:'Hiring 2 SDRs · visited pricing ×2 · fit 100', gate:'awaiting', time:'2m' },
    { id:'t2', from:'Sam Cole', co:'Oakline', title:'Head of Growth', cls:'positive', snippet:'Re: congrats on the raise — what does onboarding look like?', body:'Congrats noted, thanks! We just closed our Series A. Curious what a rollout looks like for a 90-person team.', why:'Series A 5h ago · fit 82', gate:'approved', time:'40m' },
    { id:'t3', from:'Marco Reyes', co:'Lumen Labs', title:'CMO', cls:'negative', snippet:'Not the right time — circle back in Q4.', body:'Appreciate the note but we\u2019re heads-down until Q4. Reach out then.', why:'No active signals', gate:'auto', time:'1h' },
    { id:'t4', from:'Jess Lin', co:'Brightway', title:'RevOps', cls:'ooo', snippet:'Out of office until Monday', body:'I\u2019m away until Monday with limited access to email.', why:'Auto-detected OOO · sequence paused', gate:'auto', time:'3h' },
  ],
  nimbus: [
    { id:'t1', from:'Lee Park', co:'VaultPay', title:'COO', cls:'positive', snippet:'Re: SOC2 — yes, this is timely.', body:'We\u2019re mid-audit so compliance tooling is top of mind. Let\u2019s talk.', why:'compliance_hiring · SOC2 · fit 88', gate:'awaiting', time:'12m' },
  ],
  vela: [
    { id:'t1', from:'Ira Solis', co:'Curaspan', title:'VP RevOps', cls:'positive', snippet:'Re: RevOps hire — open to a quick call.', body:'We just posted a RevOps role, so yes — a quick call could be useful.', why:'rev_ops_hire · fit 79', gate:'awaiting', time:'1h' },
  ],
}
const REPLY_CLASS = {
  positive:{ c:'var(--success)', l:'Positive' },
  negative:{ c:'var(--danger)', l:'Not now' },
  ooo:{ c:'var(--info)', l:'Out of office' },
  neutral:{ c:'var(--text-3)', l:'Neutral' },
}

// Deliverability — sending domains + warmup, suppression, health.
const DOMAINS = {
  acme:   [ { domain:'go.acme-sales.io', state:'warming', day:12, cap:50, sent:38, rep:97 }, { domain:'mail.acme-sales.io', state:'healthy', day:30, cap:200, sent:142, rep:99 }, { domain:'hi.acmereach.com', state:'warming', day:6, cap:25, sent:18, rep:95 } ],
  nimbus: [ { domain:'go.nimbus-fin.io', state:'healthy', day:30, cap:150, sent:96, rep:98 } ],
  vela:   [ { domain:'mail.vela-health.io', state:'warming', day:9, cap:40, sent:31, rep:96 } ],
}
const SUPPRESSION = [
  { addr:'pat@deadlink.com', reason:'Unsubscribed', date:'3h ago' },
  { addr:'legal@lumenlabs.com', reason:'Do-not-contact', date:'2d ago' },
  { addr:'noreply@brightway.io', reason:'Hard bounce', date:'4d ago' },
]

// Settings — capability gates, API keys, team.
const CAPABILITIES = [
  { id:'consent.email', label:'consent.email', desc:'Allow ChampMail nodes on the canvas', on:true },
  { id:'consent.voice', label:'consent.voice', desc:'Allow ChampVoice nodes (TCPA / DND checked)', on:true },
  { id:'dnd.honor',     label:'dnd.honor', desc:'Honor Do-Not-Disturb windows by timezone', on:true },
  { id:'kg.writeback',  label:'kg.writeback', desc:'Let runs write outcomes back to the graph', on:true },
  { id:'auto.voice',    label:'auto.voice', desc:'Skip human gate for hot replies (not recommended)', on:false },
]
const API_KEYS = [
  { svc:'KG store', val:'FalkorDB · falkor_live_••••8f2', ok:true },
  { svc:'Email MTA', val:'Stalwart · smtp_••••a91', ok:true },
  { svc:'Voice', val:'ElevenLabs · el_••••3dc', ok:true },
  { svc:'Signals', val:'ChampHarbinger · hb_••••77b', ok:true },
]
const TEAM = [
  { name:'Deep Sharma', role:'Owner', initial:'D', accent:'var(--accent-2)' },
  { name:'Ana Ruiz', role:'Operator', initial:'A', accent:'var(--mint-2)' },
  { name:'Tom Wells', role:'Viewer', initial:'T', accent:'var(--info)' },
]

// ── ChampMail sequence / cadence — the multi-step builder model ───────────────
// step kinds: email | wait | branch | voice ; emails carry an AI-personalized flag.
const SEQUENCE = {
  name: 'Signal-first SDR',
  enrolled: 84,
  sendWindow: 'Mon–Fri · 8am–4pm · prospect TZ',
  steps: [
    { id:'s1', kind:'email',  day:0,  title:'Signal opener', subject:'{{signal_hook}}', preview:'Grounded in the prospect\u2019s top live signal. AI-personalized per recipient.', ai:true, open:62, reply:14 },
    { id:'s2', kind:'wait',   day:2,  title:'Wait 2 days', detail:'Durable pause — Sim snapshots, resumes on reply or timeout.' },
    { id:'s3', kind:'branch', day:2,  title:'Replied?', detail:'Positive → exit to human gate · No reply → continue', branches:['Positive → Inbox gate','No reply → step 4'] },
    { id:'s4', kind:'email',  day:4,  title:'Value nudge', subject:'A quick idea for {{company}}', preview:'References the account\u2019s why-now + one concrete outcome.', ai:true, open:48, reply:9 },
    { id:'s5', kind:'wait',   day:7,  title:'Wait 3 days', detail:'Durable pause.' },
    { id:'s6', kind:'email',  day:7,  title:'Break-up', subject:'Should I close the loop?', preview:'Low-key last touch. Static copy — no signal needed.', ai:false, open:33, reply:6 },
    { id:'s7', kind:'voice',  day:9,  title:'ChampVoice (gated)', detail:'Only fires for hot intent + voice consent (TCPA/DND). Human-approved.', gated:true },
  ],
}

// ── Per-prospect AI personalization — GraphRAG grounding made visible ─────────
// body is an array of segments; segments with `g` are grounded in a KG fact id.
const PERSONALIZE = {
  dana: {
    subject: [ {t:'Scaling '}, {t:'Northwind\u2019s sales team', g:'f1'}, {t:' — a faster ramp?'} ],
    subjectAlt: 'Re: your two new SDR reqs',
    body: [
      {t:'Hi Dana, '},
      {t:'congrats on opening the VP Sales seat and the two SDR reqs', g:'f1'}, {t:' — '},
      {t:'with the team scaling fast', g:'f2'}, {t:', ramp time is usually the first thing that slips. '},
      {t:'Saw Northwind also revisited our pricing twice last week', g:'f3'}, {t:', so the timing felt right to reach out.\n\n'},
      {t:'We help SDR teams hit quota ~30% faster by grounding every touch in live buying signals — happy to share how a couple of similar teams ran it. Worth a quick 15 on Thursday?'},
    ],
    voice: [
      {t:'Hi Dana, this is Pixie calling on behalf of Champions. I saw '},
      {t:'Northwind just opened two SDR roles', g:'f1'}, {t:' and that your team is scaling — '},
      {t:'wanted to share how teams in a similar spot cut ramp time. Is now a bad moment?'},
    ],
    facts: [
      { id:'f1', label:'Hiring: VP Sales + 2 SDRs', source:'job-board signal · 2d ago', conf:0.98 },
      { id:'f2', label:'Team scaling (headcount +18% QoQ)', source:'ChampGraph enrichment', conf:0.91 },
      { id:'f3', label:'Pricing-page visit ×2', source:'web signal · 1d ago', conf:0.86 },
    ],
    fit:100, intent:'hot',
  },
  sam: {
    subject: [ {t:'Congrats on '}, {t:'Oakline\u2019s Series A', g:'g1'} ],
    subjectAlt: 'Scaling outbound post-raise?',
    body: [
      {t:'Hi Sam, '},
      {t:'congrats on closing the Series A', g:'g1'}, {t:' — exciting moment. '},
      {t:'Teams usually scale outbound right after a raise', g:'g2'}, {t:', and that\u2019s exactly where most pipelines get noisy.\n\n'},
      {t:'We keep it signal-first so your new reps only chase in-market accounts. Open to a quick look next week?'},
    ],
    voice: [
      {t:'Hi Sam — congrats on the raise at Oakline. '},
      {t:'Most teams scale outbound right after', g:'g2'}, {t:'; wanted to show how to keep it signal-first. Got a minute?'},
    ],
    facts: [
      { id:'g1', label:'Funding: Series A (5h ago)', source:'funding signal', conf:0.97 },
      { id:'g2', label:'Head of Growth · scaling motion', source:'ChampGraph role', conf:0.82 },
    ],
    fit:82, intent:'hot',
  },
}
const PERSONALIZE_FALLBACK = {
  subject: [ {t:'A quick, relevant idea'} ], subjectAlt:'Worth a look?',
  body: [ {t:'Hi there, reaching out because the timing looked right based on your account\u2019s recent activity. Worth a quick chat?'} ],
  voice: [ {t:'Hi, calling because the timing looked right — got a quick minute?'} ],
  facts: [ { id:'x', label:'No strong live signal yet', source:'enrich to improve', conf:0.4 } ], fit:60, intent:'warm',
}

Object.assign(window, { CLIENTS, GRAPHS, INTENT_COLOR, CANVAS_GRAPHS, CANVASES, TEMPLATES, FLOW, GUARD_CASES, PROSPECTS, INBOX, REPLY_CLASS, DOMAINS, SUPPRESSION, CAPABILITIES, API_KEYS, TEAM, SEQUENCE, PERSONALIZE, PERSONALIZE_FALLBACK })

/* ======== shell.jsx ======== */
/* ChampIQ (Sim) — app shell: rail nav + top bar + hash router + Copilot host. */

const NAV = [
  { id:'hub',       icon:'home',     label:'Dashboard' },
  { id:'graph',     icon:'graph',    label:'ChampGraph' },
  { id:'canvas',    icon:'grid',     label:'Canvas' },
  { id:'sequences', icon:'list',     label:'Sequences' },
  { id:'personalize', icon:'sparkle', label:'Personalize' },
  { id:'inbox',     icon:'chat',     label:'Inbox' },
  { id:'prospects', icon:'users',    label:'Prospects' },
  { id:'deliver',   icon:'mail',     label:'Deliverability' },
  { id:'analytics', icon:'layers',   label:'Analytics' },
]

function useHashRoute() {
  const [route, setRoute] = React.useState(() => (location.hash.replace('#/', '') || 'hub'))
  React.useEffect(() => {
    const on = () => setRoute(location.hash.replace('#/', '') || 'hub')
    window.addEventListener('hashchange', on)
    return () => window.removeEventListener('hashchange', on)
  }, [])
  const go = (r) => { location.hash = '#/' + r }
  return [route.split('/')[0], route.split('/')[1], go]
}

function Rail({ route, go }) {
  return (
    <div style={{ width:60, flexShrink:0, background:'var(--bg-1)', borderRight:'1px solid var(--border-1)', display:'flex', flexDirection:'column', alignItems:'center', padding:'12px 0', gap:4 }}>
      <button onClick={()=>go('hub')} title="ChampIQ" style={{ marginBottom:14, background:'transparent', border:'none', cursor:'pointer', padding:4 }}><ChampMark size={24} /></button>
      <div style={{ display:'flex', flexDirection:'column', gap:4, flex:1 }}>
        {NAV.map(n => {
          const on = route === n.id
          return (
            <button key={n.id} title={n.label} onClick={()=>go(n.id)} style={{ width:44, height:44, display:'grid', placeItems:'center', background:on?'rgba(var(--accent-2-rgb),.16)':'transparent', color:on?'var(--accent-1)':'var(--text-3)', border:'none', borderRadius:11, cursor:'pointer', boxShadow:on?'0 0 16px -4px rgba(var(--accent-2-rgb),.7), inset 0 0 0 1px rgba(var(--accent-2-rgb),.35)':'none', transition:'all .18s var(--ease-swift)' }}><Icon name={n.icon} size={19} /></button>
          )
        })}
      </div>
      <button title="Settings" onClick={()=>go('settings')} style={{ width:44, height:44, display:'grid', placeItems:'center', background:route==='settings'?'rgba(var(--accent-2-rgb),.16)':'transparent', color:route==='settings'?'var(--accent-1)':'var(--text-3)', border:'none', borderRadius:11, cursor:'pointer' }}><Icon name="settings" size={17} /></button>
    </div>
  )
}

function ClientSwitcher({ client, setClient }) {
  const [open, setOpen] = React.useState(false)
  return (
    <div style={{ position:'relative' }}>
      <button onClick={()=>setOpen(o=>!o)} style={{ display:'flex', alignItems:'center', gap:8, padding:'6px 10px', background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, cursor:'pointer', color:'var(--text-1)' }}>
        <span style={{ width:8, height:8, borderRadius:'50%', background:client.accent, boxShadow:`0 0 8px ${client.accent}` }} />
        <span style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13 }}>{client.name}</span>
        <Icon name="chevDown" size={13} stroke="var(--text-3)" />
      </button>
      {open && (
        <div style={{ position:'absolute', top:'calc(100% + 6px)', left:0, zIndex:50, background:'var(--bg-2)', border:'1px solid var(--border-2)', borderRadius:10, minWidth:220, boxShadow:'var(--shadow-lg)', overflow:'hidden', animation:'bubble-in 150ms var(--ease-spring)' }}>
          {CLIENTS.map(c => (
            <button key={c.id} onClick={()=>{ setClient(c.id); setOpen(false) }} style={{ width:'100%', display:'flex', alignItems:'center', gap:10, padding:'10px 12px', background:c.id===client.id?'var(--bg-3)':'transparent', border:'none', borderBottom:'1px solid var(--border-1)', cursor:'pointer', textAlign:'left' }}>
              <span style={{ width:8, height:8, borderRadius:'50%', background:c.accent }} />
              <div style={{ flex:1 }}><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13, color:'var(--text-1)' }}>{c.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-3)' }}>{c.prospects} prospects · {c.health}% health</div></div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function TopBar({ client, setClient, onCopilot, title }) {
  return (
    <div style={{ height:52, flexShrink:0, background:'var(--bg-1)', borderBottom:'1px solid var(--border-1)', display:'flex', alignItems:'center', padding:'0 16px', gap:14 }}>
      <div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600, color:'var(--text-1)', minWidth:120 }}>{title}</div>
      <ClientSwitcher client={client} setClient={setClient} />
      <div style={{ flex:1, display:'flex', justifyContent:'center' }}>
        <div style={{ display:'flex', alignItems:'center', gap:8, padding:'6px 12px', width:320, background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, color:'var(--text-3)', fontSize:12 }}><Icon name="search" size={13} /><span style={{ flex:1 }}>Search prospects, signals, canvases…</span><span style={{ fontFamily:'var(--font-mono)', fontSize:10, background:'var(--bg-3)', padding:'1px 6px', borderRadius:4 }}>⌘K</span></div>
      </div>
      <Button variant="pixie" size="md" icon="sparkle" onClick={onCopilot}>Ask Pixie</Button>
      <div style={{ width:30, height:30, borderRadius:'50%', background:'linear-gradient(135deg,var(--accent-2),var(--accent-3))', display:'grid', placeItems:'center', color:'#fff', fontFamily:'var(--font-display)', fontWeight:700, fontSize:12 }}>D</div>
    </div>
  )
}

window.ChampShell = { useHashRoute, Rail, TopBar }

/* ======== Copilot.jsx ======== */
/* ChampIQ (Sim) — Pixie Copilot slide-over with zerolang checked-edit guardrail. */

function GuardVerdict({ c }) {
  return (
    <div style={{ borderRadius:10, padding:'11px 13px', border:`1px solid ${c.ok?'rgba(74,222,128,.4)':'rgba(255,77,109,.4)'}`, background:c.ok?'rgba(74,222,128,.08)':'rgba(255,77,109,.08)' }}>
      <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:6 }}>
        <Icon name={c.ok?'check':'alert'} size={14} stroke={c.ok?'var(--success)':'var(--danger)'} />
        <span style={{ fontFamily:'var(--font-mono)', fontSize:10, letterSpacing:'.08em', textTransform:'uppercase', color:c.ok?'var(--success)':'var(--danger)' }}>{c.ok?'edit accepted':'edit rejected'}</span>
      </div>
      <div style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.5 }}>{c.msg}</div>
      <pre style={{ margin:'8px 0 0', padding:'7px 9px', borderRadius:6, background:'var(--bg-0)', border:'1px solid var(--border-1)', fontFamily:'var(--font-mono)', fontSize:10.5, color:c.ok?'var(--success)':'var(--danger)', overflowX:'auto' }}>{c.code}</pre>
    </div>
  )
}

function Copilot({ open, onClose }) {
  const [tab, setTab] = React.useState('chat')
  const [verdict, setVerdict] = React.useState(null)
  const [thread, setThread] = React.useState([
    { who:'pixie', text:'I can build or edit any canvas for this client — every change runs through the zerolang checked-edit guardrail before it touches the graph.' },
  ])
  const [draft, setDraft] = React.useState('')

  function send(text) {
    const t = text ?? draft
    if (!t.trim()) return
    setThread(th => [...th, { who:'you', text:t }])
    setDraft('')
    setTimeout(() => {
      setThread(th => [...th, { who:'pixie', text:'On it — I drafted a checked edit. Review the guardrail verdict in the Checked-edit tab.', action:'View checked edit' }])
      setTab('guard'); setVerdict(GUARD_CASES[0])
    }, 500)
  }

  if (!open) return null
  return (
    <>
      <div onClick={onClose} style={{ position:'absolute', inset:0, background:'rgba(7,9,18,.55)', zIndex:60, animation:'fade-in 160ms ease' }} />
      <div className="ciq-slide-in" style={{ position:'absolute', top:0, right:0, bottom:0, width:380, zIndex:61, background:'var(--bg-1)', borderLeft:'1px solid var(--border-2)', display:'flex', flexDirection:'column', boxShadow:'var(--shadow-lg)' }}>
        <div style={{ padding:'14px 16px', borderBottom:'1px solid var(--border-1)', display:'flex', alignItems:'center', gap:10 }}>
          <Pixie pose="point" size={48} ambient={false} />
          <div style={{ flex:1 }}>
            <div className="t-pixel" style={{ color:'var(--mint-2)' }}>PIXIE · COPILOT</div>
            <div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600, color:'var(--text-1)' }}>Build with guardrails</div>
          </div>
          <button onClick={onClose} style={{ width:28, height:28, display:'grid', placeItems:'center', background:'transparent', border:'none', color:'var(--text-3)', cursor:'pointer', borderRadius:6 }}><Icon name="x" size={15} /></button>
        </div>

        <div style={{ display:'flex', padding:'8px 10px', gap:6, borderBottom:'1px solid var(--border-1)' }}>
          {[['chat','Chat','sparkle'],['guard','Checked-edit','check']].map(([id,label,ic]) => {
            const on = tab===id
            return <button key={id} onClick={()=>setTab(id)} style={{ flex:1, display:'flex', alignItems:'center', justifyContent:'center', gap:6, padding:'6px 0', borderRadius:7, border:'none', cursor:'pointer', fontFamily:'var(--font-display)', fontWeight:600, fontSize:12, background:on?'var(--bg-3)':'transparent', color:on?'var(--accent-1)':'var(--text-3)' }}><Icon name={ic} size={13} />{label}</button>
          })}
        </div>

        {tab === 'chat' ? (
          <>
            <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:14, display:'flex', flexDirection:'column', gap:14 }}>
              {thread.map((m,i) => {
                const me = m.who==='you'
                return (
                  <div key={i} style={{ display:'flex', flexDirection:'column', alignItems:me?'flex-end':'flex-start', gap:6 }}>
                    {!me && <div className="t-pixel" style={{ color:'var(--mint-2)', marginLeft:2 }}>PIXIE</div>}
                    <div style={{ maxWidth:'88%', padding:'9px 12px', borderRadius:me?'12px 12px 4px 12px':'12px 12px 12px 4px', background:me?'linear-gradient(180deg,var(--accent-2),var(--accent-3))':'var(--bg-2)', border:me?'1px solid var(--accent-3)':'1px solid var(--border-1)', color:me?'#fff':'var(--text-2)', fontSize:13, lineHeight:1.5, animation:'bubble-in 200ms var(--ease-spring)' }}>
                      {m.text}
                      {m.action && <div style={{ marginTop:8 }}><Button variant="pixie" size="sm" icon="check" onClick={()=>setTab('guard')}>{m.action}</Button></div>}
                    </div>
                  </div>
                )
              })}
            </div>
            <div style={{ padding:'8px 12px 0', display:'flex', gap:6, flexWrap:'wrap' }}>
              {['Add a voice step for hot replies','Insert a 2-day wait','Branch on positive intent'].map(s=>(
                <button key={s} onClick={()=>send(s)} style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--accent-1)', background:'rgba(var(--accent-2-rgb),.1)', border:'1px solid rgba(var(--accent-2-rgb),.25)', borderRadius:999, padding:'4px 9px', cursor:'pointer' }}>{s}</button>
              ))}
            </div>
            <div style={{ padding:'10px 12px', borderTop:'1px solid var(--border-1)', display:'flex', gap:8, alignItems:'center', marginTop:8 }}>
              <Input placeholder="Describe a change — I'll build it safely…" value={draft} onChange={e=>setDraft(e.target.value)} onKeyDown={e=>{ if(e.key==='Enter') send() }} />
              <Button variant="primary" size="md" icon="send" onClick={()=>send()} style={{ flexShrink:0 }} />
            </div>
          </>
        ) : (
          <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:14, display:'flex', flexDirection:'column', gap:12 }}>
            <div style={{ fontSize:12.5, color:'var(--text-3)', lineHeight:1.5 }}>Every Copilot edit is a checked graph-edit: <b style={{ color:'var(--text-2)' }}>graphHash precondition</b> + <b style={{ color:'var(--text-2)' }}>expect</b> checks + <b style={{ color:'var(--text-2)' }}>edge connectivity</b> + a <b style={{ color:'var(--text-2)' }}>compliance capability gate</b>. Try one:</div>
            <div style={{ display:'flex', flexDirection:'column', gap:6 }}>
              {GUARD_CASES.map(c => (
                <button key={c.id} onClick={()=>setVerdict(c)} style={{ display:'flex', alignItems:'center', gap:8, padding:'9px 11px', borderRadius:8, border:`1px solid ${verdict&&verdict.id===c.id?(c.ok?'var(--success)':'var(--danger)'):'var(--border-1)'}`, background:'var(--bg-2)', color:'var(--text-2)', cursor:'pointer', textAlign:'left', fontSize:12, fontFamily:'var(--font-body)' }}>
                  <Icon name={c.ok?'check':'alert'} size={13} stroke={c.ok?'var(--success)':'var(--danger)'} />{c.label}
                </button>
              ))}
            </div>
            {verdict && <GuardVerdict c={verdict} />}
          </div>
        )}
      </div>
    </>
  )
}

window.Copilot = Copilot

/* ======== Login.jsx ======== */
/* ChampIQ (Sim) — Login / Workspace gate. Pick the client (KG namespace) to enter. */

function LoginScreen({ onEnter }) {
  const [sel, setSel] = React.useState('acme')
  return (
    <div style={{ width:'100%', height:'100%', display:'flex', background:'var(--bg-0)', color:'var(--text-1)', fontFamily:'var(--font-body)', overflow:'hidden' }}>
      {/* Left — brand / thesis */}
      <div className="bg-grid" style={{ flex:1, position:'relative', display:'flex', flexDirection:'column', justifyContent:'center', padding:'0 64px', borderRight:'1px solid var(--border-1)', overflow:'hidden' }}>
        <div style={{ position:'absolute', inset:'-30% 30% auto -10%', height:520, background:'radial-gradient(closest-side, rgba(var(--accent-2-rgb),.18), transparent)', filter:'blur(20px)', pointerEvents:'none' }} />
        <div style={{ position:'relative' }}>
          <Wordmark size={30} />
          <div className="t-pixel" style={{ color:'var(--mint-2)', margin:'30px 0 14px' }}>SDR AUTOMATION ON THE SIM RUNTIME</div>
          <h1 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:40, fontWeight:700, letterSpacing:'-.03em', lineHeight:1.05, maxWidth:460 }}>The SDR brain<br/>on top of Sim.</h1>
          <p style={{ color:'var(--text-2)', fontSize:15, lineHeight:1.6, maxWidth:440, marginTop:16 }}>Engine = commodity. Moat = the graph + the guardrail. Every client owns an isolated knowledge graph; outreach is grounded in real account signals; and no compliance-sensitive node ships without its consent capability.</p>
          <div style={{ display:'flex', flexWrap:'wrap', gap:8, marginTop:26, maxWidth:460 }}>
            {['Per-client KG','GraphRAG personalization','zerolang guardrail','Consent-gated voice','Owned deliverability'].map(c=>(
              <span key={c} style={{ display:'inline-flex', alignItems:'center', gap:7, padding:'6px 12px', borderRadius:999, border:'1px solid var(--border-2)', background:'var(--bg-1)', fontSize:12.5, color:'var(--text-2)' }}><span style={{ width:6, height:6, borderRadius:'50%', background:'var(--accent-2)' }} />{c}</span>
            ))}
          </div>
          <div style={{ position:'absolute', right:-12, bottom:-30 }}><Pixie pose="lean" size={120} /></div>
        </div>
      </div>

      {/* Right — sign in + workspace */}
      <div style={{ width:440, flexShrink:0, display:'flex', flexDirection:'column', justifyContent:'center', padding:'0 44px' }}>
        <div style={{ fontFamily:'var(--font-display)', fontSize:22, fontWeight:600, letterSpacing:'-.01em' }}>Sign in</div>
        <div style={{ fontSize:13, color:'var(--text-3)', marginTop:4, marginBottom:22 }}>Then choose the marketing client to load its graph.</div>

        <label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:6, fontFamily:'var(--font-mono)', letterSpacing:'.08em', textTransform:'uppercase' }}>Email</label>
        <Input value="deep@championsgroup.io" readOnly icon="user" style={{ marginBottom:14 }} />
        <label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:6, fontFamily:'var(--font-mono)', letterSpacing:'.08em', textTransform:'uppercase' }}>Password</label>
        <Input value="••••••••••" readOnly type="password" style={{ marginBottom:22 }} />

        <label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:8, fontFamily:'var(--font-mono)', letterSpacing:'.08em', textTransform:'uppercase' }}>Workspace · KG namespace</label>
        <div style={{ display:'flex', flexDirection:'column', gap:8, marginBottom:24 }}>
          {CLIENTS.map(c => {
            const on = sel===c.id
            return (
              <button key={c.id} onClick={()=>setSel(c.id)} style={{ display:'flex', alignItems:'center', gap:12, padding:'11px 13px', borderRadius:10, cursor:'pointer', textAlign:'left', background:on?'rgba(var(--accent-2-rgb),.1)':'var(--bg-1)', border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, transition:'all .15s' }}>
                <span style={{ width:10, height:10, borderRadius:'50%', background:c.accent, boxShadow:on?`0 0 8px ${c.accent}`:'none', flexShrink:0 }} />
                <div style={{ flex:1 }}><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13.5 }}>{c.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)' }}>client:{c.id} · {c.prospects} prospects</div></div>
                {on && <Icon name="check" size={16} stroke="var(--accent-2)" />}
              </button>
            )
          })}
        </div>

        <Button variant="primary" size="lg" icon="arrow_right" onClick={()=>onEnter(sel)} style={{ width:'100%', justifyContent:'center' }}>Enter ChampIQ</Button>
        <div style={{ display:'flex', alignItems:'center', gap:6, justifyContent:'center', marginTop:16, fontSize:11.5, color:'var(--text-4)', fontFamily:'var(--font-mono)' }}><Icon name="check" size={12} stroke="var(--success)" />SSO · SAML available · workspace branding applied</div>
      </div>
    </div>
  )
}

window.LoginScreen = LoginScreen

/* ======== Hub.jsx ======== */
/* ChampIQ (Sim) — Hub page. */

function StatTile({ value, label, color }) {
  return (
    <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'16px 18px' }}>
      <div style={{ fontFamily:'var(--font-display)', fontSize:26, fontWeight:700, color, letterSpacing:'-.02em' }}>{value}</div>
      <div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-3)', marginTop:3, textTransform:'uppercase', letterSpacing:'.1em' }}>{label}</div>
    </div>
  )
}

function SectionHeader({ title, sub, action }) {
  return (
    <div style={{ display:'flex', alignItems:'flex-end', justifyContent:'space-between', marginBottom:14 }}>
      <div>
        <h3 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:18, fontWeight:600, color:'var(--text-1)', letterSpacing:'-.01em' }}>{title}</h3>
        {sub && <div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-3)', marginTop:3, letterSpacing:'.14em', textTransform:'uppercase' }}>{sub}</div>}
      </div>
      {action}
    </div>
  )
}

function HubCanvasCard({ canvas, onOpen, i }) {
  const [over,setOver] = React.useState(false)
  const client = CLIENTS.find(c=>c.id===canvas.client)
  const accent = client?.accent ?? 'var(--accent-2)'
  return (
    <div onClick={onOpen} onMouseEnter={()=>setOver(true)} onMouseLeave={()=>setOver(false)} style={{ background:'var(--bg-1)', border:'1px solid', borderColor:over?accent:'var(--border-1)', borderRadius:12, padding:16, cursor:'pointer', position:'relative', overflow:'hidden', transition:'transform .18s var(--ease-swift), border-color .18s, box-shadow .18s', transform:over?'translateY(-2px)':'none', boxShadow:over?`0 8px 28px -10px ${accent}88`:'none', animation:`hub-card-in 420ms var(--ease-spring) ${i*30}ms backwards` }}>
      <div style={{ position:'absolute', top:0, left:0, right:0, height:3, background:accent }} />
      <div style={{ display:'flex', alignItems:'flex-start', gap:12 }}>
        <div style={{ width:38, height:38, borderRadius:9, flexShrink:0, background:`color-mix(in oklch, ${accent} 22%, transparent)`, color:accent, display:'grid', placeItems:'center' }}><Icon name="layers" size={18} /></div>
        <div style={{ flex:1, minWidth:0 }}>
          <div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600, color:'var(--text-1)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{canvas.name}</div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, letterSpacing:'.06em', color:'var(--text-3)', marginTop:2 }}>{canvas.nodes} nodes · {canvas.updated}</div>
        </div>
        <Badge status={canvas.status} pulse={canvas.status==='running'} />
      </div>
    </div>
  )
}

const SIGNALS_FEED = [
  { dot:'var(--success)', text:'Northwind posted "VP Sales" — Dana Okafor matched (fit 100)', time:'2m' },
  { dot:'var(--mint-2)', text:'Pixie drafted a signal-personalized opener for Oakline', time:'14m' },
  { dot:'var(--info)', text:'VaultPay raised Series C — 4 new prospects enriched', time:'1h' },
  { dot:'var(--warn)', text:'Voice screener paused — awaiting consent capability', time:'2h' },
  { dot:'var(--success)', text:'Reply: Priya Nair (Vela) — positive, routed to human', time:'3h' },
]

function HubPage({ client, onOpenCanvas, onNewCanvas, onCopilot, go }) {
  const canvases = CANVASES.filter(c => c.client === client.id)
  const list = canvases.length ? canvases : CANVASES.slice(0,2)
  return (
    <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
      <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'28px 32px' }}>
        {/* Pixie briefing — signal-first */}
        <div style={{ background:'linear-gradient(135deg, rgba(var(--accent-2-rgb),.12), rgba(var(--mint-2-rgb),.06) 80%)', border:'1px solid rgba(var(--accent-2-rgb),.25)', borderRadius:16, padding:24, display:'flex', gap:22, alignItems:'center', position:'relative', overflow:'hidden' }}>
          <div style={{ position:'absolute', inset:0, opacity:.25, pointerEvents:'none', backgroundImage:'radial-gradient(circle at 1px 1px, rgba(255,255,255,.18) 1px, transparent 0)', backgroundSize:'14px 14px' }} />
          <div style={{ flexShrink:0, position:'relative' }}><Pixie pose="point" size={118} /></div>
          <div style={{ flex:1, position:'relative' }}>
            <div className="t-pixel" style={{ color:'var(--mint-2)', marginBottom:6 }}>PIXIE · BRIEFING · {client.name.toUpperCase()}</div>
            <div style={{ fontFamily:'var(--font-display)', fontSize:22, fontWeight:600, color:'var(--text-1)', marginBottom:6, letterSpacing:'-.01em' }}>3 hot signals broke overnight.</div>
            <div style={{ fontSize:14, color:'var(--text-2)', lineHeight:1.55, maxWidth:560 }}>Northwind is hiring a VP Sales and visited pricing twice — <b style={{ color:'var(--text-1)' }}>Dana Okafor</b> scored fit 100. Want me to build a signal-first sequence for the new matches?</div>
            <div style={{ display:'flex', gap:8, marginTop:16 }}>
              <Button variant="pixie" size="md" icon="sparkle" onClick={onCopilot}>Build it with me</Button>
              <Button variant="ghost" size="md" onClick={()=>onOpenCanvas(list[0])}>Open the canvas</Button>
            </div>
          </div>
        </div>

        {/* Moat strip */}
        <div style={{ display:'grid', gridTemplateColumns:'repeat(3,1fr)', gap:12, marginTop:28 }}>
          {[['graph','Per-client KG','Isolated graph · grounded personalization','var(--accent-2)'],['inbox','zerolang guardrail','Every AI edit checked before it ships','var(--mint-2)'],['deliver','Owned deliverability','Per-client domains · warmup · suppression','var(--info)']].map(([dest,t,d,c])=>(
            <button key={t} onClick={()=>go(dest)} style={{ textAlign:'left', display:'flex', gap:12, alignItems:'flex-start', padding:'14px 16px', borderRadius:12, background:'var(--bg-1)', border:'1px solid var(--border-1)', cursor:'pointer' }}>
              <div style={{ width:30, height:30, borderRadius:8, flexShrink:0, background:`color-mix(in oklch, ${c} 20%, transparent)`, color:c, display:'grid', placeItems:'center' }}><Icon name={dest==='graph'?'graph':dest==='inbox'?'check':'mail'} size={15} /></div>
              <div><div style={{ fontFamily:'var(--font-display)', fontSize:13, fontWeight:600, color:'var(--text-1)' }}>{t}</div><div style={{ fontSize:11.5, color:'var(--text-3)', marginTop:2, lineHeight:1.4 }}>{d}</div></div>
            </button>
          ))}
        </div>

        <div style={{ marginTop:28 }}>
          <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12, marginBottom:28 }}>
            <StatTile value={client.prospects} label="Prospects in graph" color={client.accent} />
            <StatTile value={client.signals} label="Live signals" color="#F97316" />
            <StatTile value={client.replies} label="Positive replies · 7d" color="var(--success)" />
            <StatTile value={client.health + '%'} label="Account health" color="var(--info)" />
          </div>

          <div style={{ marginBottom:32 }}>
            <SectionHeader title="Active canvases" sub={`${list.length} for ${client.name}`} action={<Button variant="ghost" size="sm" icon="plus" onClick={()=>onNewCanvas(TEMPLATES[0])}>New canvas</Button>} />
            <div style={{ display:'grid', gridTemplateColumns:'repeat(3,1fr)', gap:14 }}>
              {list.map((c,i)=><HubCanvasCard key={c.id} canvas={c} i={i} onOpen={()=>onOpenCanvas(c)} />)}
            </div>
          </div>

          <SectionHeader title="Start from a template" sub="Curated by Pixie" />
          <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12 }}>
            {TEMPLATES.map(t => {
              const [over,setOver]=React.useState(false)
              return (
                <div key={t.title} onClick={()=>onNewCanvas(t)} onMouseEnter={()=>setOver(true)} onMouseLeave={()=>setOver(false)} style={{ background:over?'var(--bg-2)':'var(--bg-1)', border:`1px dashed ${over?'var(--accent-2)':'var(--border-1)'}`, borderRadius:10, padding:14, cursor:'pointer', transition:'all .18s var(--ease-swift)' }}>
                  <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}><span style={{ color:t.color }}><Icon name="bolt" size={13} /></span><Tag color={t.color}>{t.tag}</Tag></div>
                  <div style={{ fontFamily:'var(--font-display)', fontSize:13, fontWeight:600, color:'var(--text-1)', marginBottom:4 }}>{t.title}</div>
                  <div style={{ fontSize:11, color:'var(--text-3)', lineHeight:1.5 }}>{t.desc}</div>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* Signal feed */}
      <div style={{ width:300, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', padding:'24px 20px', overflowY:'auto' }}>
        <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:16 }}><Icon name="bolt" size={15} stroke="#F97316" /><div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600, color:'var(--text-1)' }}>Signal feed</div></div>
        <div style={{ display:'flex', flexDirection:'column' }}>
          {SIGNALS_FEED.map((a,i)=>(
            <div key={i} style={{ display:'flex', gap:10, padding:'11px 0', borderBottom:i<SIGNALS_FEED.length-1?'1px solid var(--border-1)':'none' }}>
              <span style={{ width:6, height:6, borderRadius:'50%', background:a.dot, marginTop:5, flexShrink:0, boxShadow:`0 0 6px ${a.dot}` }} />
              <div style={{ flex:1 }}><div style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.45 }}>{a.text}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)', marginTop:3 }}>{a.time} ago</div></div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

window.HubPage = HubPage

/* ======== Canvas.jsx ======== */
/* ChampIQ (Sim) — Canvas cockpit: the signal-first SDR DAG with durable run. */

const C_PALETTE = [
  { group:'Signals', items:[['harbinger','ChampHarbinger'],['webhook','Webhook'],['cron','Cron']] },
  { group:'Data / AI', items:[['champgraph','ChampGraph'],['llm','Personalize'],['classifier','Reply classifier']] },
  { group:'Control', items:[['if','If / Branch'],['split','Split A/B'],['wait','Wait / Delay']] },
  { group:'Human', items:[['approval','Approval gate']] },
  { group:'Channels', items:[['champmail','ChampMail'],['champvoice','ChampVoice']] },
]

function CanvasPalette() {
  return (
    <aside style={{ width:210, flexShrink:0, background:'var(--bg-1)', borderRight:'1px solid var(--border-1)', overflowY:'auto' }}>
      <div style={{ padding:'10px 12px 8px', borderBottom:'1px solid var(--border-1)', fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)' }}>Blocks</div>
      {C_PALETTE.map(g => (
        <div key={g.group} style={{ padding:'10px 12px' }}>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:6 }}>{g.group}</div>
          <div style={{ display:'flex', flexDirection:'column', gap:4 }}>
            {g.items.map(([kind,label]) => {
              const color = KIND_COLOR[kind] ?? 'var(--accent-2)'
              const [over,setOver]=React.useState(false)
              const icon = ({harbinger:'search',webhook:'webhook',cron:'clock',champgraph:'graph',llm:'sparkle',classifier:'branch',if:'if_node',split:'branch',wait:'clock',approval:'user',champmail:'mail',champvoice:'voice'})[kind]
              return (
                <div key={kind} onMouseEnter={()=>setOver(true)} onMouseLeave={()=>setOver(false)} style={{ display:'flex', alignItems:'center', gap:8, padding:'6px 8px', borderRadius:8, background:over?'var(--bg-3)':'var(--bg-2)', border:`1px solid ${over?color:'var(--border-1)'}`, borderLeft:`3px solid ${color}`, cursor:'grab', transition:'all .15s' }}>
                  <span style={{ color, display:'grid', placeItems:'center' }}><Icon name={icon} size={13} /></span>
                  <span style={{ fontSize:11, fontFamily:'var(--font-display)', fontWeight:500, color:'var(--text-2)' }}>{label}</span>
                </div>
              )
            })}
          </div>
        </div>
      ))}
    </aside>
  )
}

function edgePath(x1,y1,x2,y2){ const dx=Math.max(40,Math.abs(x2-x1)*0.5); return `M ${x1},${y1} C ${x1+dx},${y1} ${x2-dx},${y2} ${x2},${y2}` }

function CanvasBoard({ nodes, edges, selected, onSelect }) {
  const NW=200, NH=84
  const W = Math.max(...nodes.map(n=>n.x)) + 320
  const H = Math.max(...nodes.map(n=>n.y)) + 200
  return (
    <div className="scroll" style={{ flex:1, overflow:'auto', background:'var(--bg-0)' }}>
      <div className="bg-grid" style={{ position:'relative', width:W, height:H, minWidth:'100%', minHeight:'100%' }}>
        <svg style={{ position:'absolute', inset:0, width:W, height:H, pointerEvents:'none' }}>
          {edges.map((e,i)=>{
            const a=nodes.find(n=>n.id===e.from), b=nodes.find(n=>n.id===e.to)
            if(!a||!b) return null
            const x1=a.x+NW, y1=a.y+NH/2, x2=b.x, y2=b.y+NH/2
            const active=e.state==='active', done=e.state==='done'
            return <path key={i} d={edgePath(x1,y1,x2,y2)} fill="none" stroke={active?'var(--accent-2)':done?'var(--success)':'var(--border-2)'} strokeWidth={active?2.5:2} strokeDasharray={e.state==='waiting'||e.state==='idle'?'5 5':active?'8 6':'none'} style={active?{animation:'dash-flow 0.6s linear infinite'}:undefined} />
          })}
        </svg>
        {nodes.map(n=>(
          <div key={n.id} style={{ position:'absolute', left:n.x, top:n.y }} onClick={()=>onSelect(n.id)}>
            <NodeCard kind={n.kind} label={n.label} status={n.status} selected={selected===n.id} />
          </div>
        ))}
      </div>
    </div>
  )
}

function RunTimeline({ steps, cursor, paused }) {
  return (
    <div style={{ height:62, flexShrink:0, background:'var(--bg-1)', borderTop:'1px solid var(--border-1)', display:'flex', alignItems:'center', gap:0, padding:'0 16px', overflowX:'auto' }}>
      <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginRight:16, flexShrink:0 }}>Run</div>
      {steps.map((s,i)=>{
        const state = i<cursor ? 'done' : i===cursor ? (paused?'paused':'running') : 'idle'
        const col = state==='done'?'var(--success)':state==='running'?'var(--warn)':state==='paused'?'var(--info)':'var(--text-4)'
        return (
          <div key={i} style={{ display:'flex', alignItems:'center', flexShrink:0 }}>
            <div style={{ display:'flex', alignItems:'center', gap:7, padding:'5px 10px', borderRadius:8, background:state==='idle'?'transparent':`color-mix(in oklch, ${col} 12%, transparent)`, border:`1px solid ${state==='idle'?'var(--border-1)':`color-mix(in oklch, ${col} 35%, transparent)`}` }}>
              <span style={{ width:7, height:7, borderRadius:'50%', background:col, boxShadow:state!=='idle'?`0 0 6px ${col}`:'none', animation:state==='running'?'glow-pulse 1s ease-in-out infinite':'none' }} />
              <span style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:state==='idle'?'var(--text-4)':'var(--text-2)', whiteSpace:'nowrap' }}>{s.t}{state==='paused'?' · snapshot':''}</span>
            </div>
            {i<steps.length-1 && <Icon name="chevRight" size={12} stroke="var(--text-4)" style={{ margin:'0 2px' }} />}
          </div>
        )
      })}
    </div>
  )
}

function Inspector({ node, onClose, onCopilot }) {
  if (!node) return (
    <div style={{ width:300, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', display:'flex', flexDirection:'column', alignItems:'center', justifyContent:'center', gap:16, padding:24, textAlign:'center' }}>
      <Pixie pose="lean" size={92} />
      <div style={{ fontSize:13, color:'var(--text-3)', lineHeight:1.5 }}>Select a node to inspect its config, KG grounding, and last output — or ask Pixie to edit the graph.</div>
      <Button variant="pixie" size="md" icon="sparkle" onClick={onCopilot}>Edit with Pixie</Button>
    </div>
  )
  const color = KIND_COLOR[node.kind] ?? 'var(--accent-2)'
  const cfg = ({ harbinger:['ICP','Acme SaaS · 50–200 · hiring'], champgraph:['Namespace','client:acme (isolated)'], llm:['Model','claude-sonnet-4.6'], champmail:['Sequence','Q3 signal-first · 4 steps'], wait:['Resume on','reply webhook · max 3 days'], classifier:['Classes','positive · negative · neutral'], approval:['Gate','human · high-value calls'], champvoice:['Capability','consent.voice ✓ · DND-checked'], webhook:['Source','inbound route'], if:['Condition','intent == positive'] })[node.kind] ?? ['Config','—']
  return (
    <div className="ciq-slide-in" style={{ width:300, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', display:'flex', flexDirection:'column' }}>
      <div style={{ padding:'14px 16px', borderBottom:'1px solid var(--border-1)', display:'flex', alignItems:'center', gap:10 }}>
        <div style={{ width:32, height:32, borderRadius:8, background:`color-mix(in oklch, ${color} 22%, transparent)`, color, display:'grid', placeItems:'center' }}><Icon name={({harbinger:'search',champgraph:'graph',llm:'sparkle',champmail:'mail',wait:'clock',classifier:'branch',approval:'user',champvoice:'voice',webhook:'webhook',if:'if_node'})[node.kind]||'layers'} size={16} /></div>
        <div style={{ flex:1 }}><div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)' }}>{node.kind}</div><div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600, color:'var(--text-1)' }}>{node.label}</div></div>
        <button onClick={onClose} style={{ width:26, height:26, display:'grid', placeItems:'center', background:'transparent', border:'none', color:'var(--text-3)', cursor:'pointer', borderRadius:6 }}><Icon name="x" size={14} /></button>
      </div>
      <div style={{ padding:16, display:'flex', flexDirection:'column', gap:16, overflowY:'auto' }}>
        <div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:8 }}>Configuration</div>
          <label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:4 }}>{cfg[0]}</label>
          <Input value={cfg[1]} readOnly />
        </div>
        {node.kind==='champvoice' && <div style={{ display:'flex', alignItems:'center', gap:8, padding:'10px 12px', borderRadius:8, background:'rgba(91,192,255,.08)', border:'1px solid rgba(91,192,255,.3)' }}><Icon name="check" size={14} stroke="var(--info)" /><span style={{ fontSize:12, color:'var(--text-2)' }}>Compliance gate passed — consent.voice declared.</span></div>}
        <div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:8 }}>Last output</div>
          <pre style={{ margin:0, padding:'10px 12px', borderRadius:8, background:'var(--bg-0)', border:'1px solid var(--border-1)', fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', lineHeight:1.6, overflowX:'auto' }}>{`{ "records": 142,\n  "status": "ok",\n  "ms": 1280 }`}</pre>
        </div>
        <Button variant="pixie" size="md" icon="sparkle" onClick={onCopilot}>Ask Pixie to edit</Button>
      </div>
    </div>
  )
}

function CanvasPage({ canvas, onCopilot }) {
  const [nodes, setNodes] = React.useState(canvas.graph.nodes.map(n=>({...n})))
  const [edges, setEdges] = React.useState(canvas.graph.edges.map(e=>({...e})))
  const [selected, setSelected] = React.useState(null)
  const [running, setRunning] = React.useState(false)
  const [cursor, setCursor] = React.useState(-1)
  const [paused, setPaused] = React.useState(false)
  const selNode = nodes.find(n=>n.id===selected)

  React.useEffect(()=>{ setNodes(canvas.graph.nodes.map(n=>({...n}))); setEdges(canvas.graph.edges.map(e=>({...e}))); setCursor(-1); setRunning(false); setPaused(false) }, [canvas])

  function runAll() {
    if (running) return
    const seq = canvas.graph.nodes              // stable sequence for the run
    setRunning(true); setPaused(false)
    setNodes(seq.map(n=>({...n, status:'idle'})))
    setEdges(canvas.graph.edges.map(e=>({...e, state:'idle'})))
    let i = 0
    const step = () => {
      if (i >= seq.length) { setRunning(false); setPaused(false); setCursor(seq.length); return }
      const n = seq[i]
      setCursor(i)
      const isWait = n.kind === 'wait'
      if (isWait) setPaused(true)
      setNodes(ns=>ns.map(x=>x.id===n.id?{...x, status: isWait?'waiting':'running'}:x))
      setEdges(es=>es.map(e=>e.to===n.id?{...e, state:'active'}:e))
      setTimeout(()=>{                            // durable wait pauses ~3× longer (snapshot/resume)
        setPaused(false)
        setNodes(ns=>ns.map(x=>x.id===n.id?{...x, status:'success'}:x))
        setEdges(es=>es.map(e=>e.to===n.id?{...e, state:'done'}:e))
        i++; step()
      }, isWait ? 1700 : 560)
    }
    step()
  }

  return (
    <div style={{ flex:1, display:'flex', flexDirection:'column', minHeight:0 }}>
      {/* Sub-toolbar */}
      <div style={{ height:46, flexShrink:0, background:'var(--bg-1)', borderBottom:'1px solid var(--border-1)', display:'flex', alignItems:'center', padding:'0 14px', gap:10 }}>
        <Icon name="layers" size={15} stroke="var(--accent-1)" />
        <span style={{ fontFamily:'var(--font-display)', fontSize:13, fontWeight:600, color:'var(--text-1)' }}>{canvas.name}</span>
        <Tag color="var(--text-4)">Auto-save</Tag>
        <span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>Sim runtime · durable</span>
        <div style={{ flex:1 }} />
        <Button variant="ghost" size="md" icon="save">Save</Button>
        <Button variant="secondary" size="md">Activate</Button>
        <Button variant="primary" size="md" icon="play" onClick={runAll} disabled={running}>{running?(paused?'Paused · snapshot':'Running…'):'Run All'}</Button>
      </div>
      <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
        <CanvasPalette />
        <CanvasBoard nodes={nodes} edges={edges} selected={selected} onSelect={setSelected} />
        <Inspector node={selNode} onClose={()=>setSelected(null)} onCopilot={onCopilot} />
      </div>
      <RunTimeline steps={canvas.graph.nodes} cursor={cursor} paused={paused} />
    </div>
  )
}

window.CanvasPage = CanvasPage

/* ======== Graph.jsx ======== */
/* ChampIQ (Sim) — ChampGraph: per-client knowledge graph dashboard (the moat). */

function FitBar({ fit, intent }) {
  const col = INTENT_COLOR[intent]
  return (
    <div style={{ display:'flex', alignItems:'center', gap:8 }}>
      <div style={{ flex:1, height:7, borderRadius:999, background:'var(--bg-3)', overflow:'hidden' }}><div style={{ width:fit+'%', height:'100%', background:col, borderRadius:999 }} /></div>
      <span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:col, width:54, textAlign:'right' }}>fit {fit}</span>
    </div>
  )
}

// Stakeholder map — company centre, prospects orbiting, reportsTo edges.
function StakeholderMap({ graph, selected, onSelect }) {
  const W=560, H=360, cx=W/2, cy=H/2
  // place companies, then prospects around their company
  const companies = graph.companies
  const compPos = {}
  companies.forEach((c,i) => {
    const ang = (i / companies.length) * Math.PI * 2 - Math.PI/2
    const r = companies.length===1 ? 0 : 120
    compPos[c.id] = { x: cx + Math.cos(ang)*r, y: cy + Math.sin(ang)*r }
  })
  const proPos = {}
  companies.forEach(c => {
    const kids = graph.prospects.filter(p=>p.company===c.id)
    kids.forEach((p,j) => {
      const ang = (j / Math.max(kids.length,1)) * Math.PI * 2 + 0.6
      const r = 92
      proPos[p.id] = { x: compPos[c.id].x + Math.cos(ang)*r, y: compPos[c.id].y + Math.sin(ang)*r }
    })
  })
  return (
    <div style={{ position:'relative', width:'100%', height:H, background:'var(--bg-0)', borderRadius:12, border:'1px solid var(--border-1)', overflow:'hidden' }} className="bg-grid">
      <svg style={{ position:'absolute', inset:0, width:'100%', height:H, pointerEvents:'none' }} viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="xMidYMid meet">
        {/* company → prospect (works_at) */}
        {graph.prospects.map(p => { const a=compPos[p.company], b=proPos[p.id]; if(!a||!b) return null; return <line key={'w'+p.id} x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="var(--border-2)" strokeWidth="1.5" /> })}
        {/* reportsTo */}
        {graph.prospects.filter(p=>p.reportsTo).map(p => { const a=proPos[p.reportsTo], b=proPos[p.id]; if(!a||!b) return null; return <line key={'r'+p.id} x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="var(--accent-3)" strokeWidth="1.5" strokeDasharray="4 4" /> })}
      </svg>
      {companies.map(c => (
        <div key={c.id} style={{ position:'absolute', left:`${compPos[c.id].x/W*100}%`, top:`${compPos[c.id].y/H*100}%`, transform:'translate(-50%,-50%)', display:'flex', flexDirection:'column', alignItems:'center', gap:4 }}>
          <div style={{ width:46, height:46, borderRadius:12, background:'var(--bg-2)', border:'1px solid var(--border-2)', display:'grid', placeItems:'center', color:'var(--text-2)' }}><Icon name="db" size={18} /></div>
          <span style={{ fontFamily:'var(--font-display)', fontSize:11, fontWeight:600, color:'var(--text-1)' }}>{c.name}</span>
        </div>
      ))}
      {graph.prospects.map(p => {
        const col = INTENT_COLOR[p.intent], on = selected===p.id
        return (
          <button key={p.id} onClick={()=>onSelect(p.id)} style={{ position:'absolute', left:`${proPos[p.id].x/W*100}%`, top:`${proPos[p.id].y/H*100}%`, transform:'translate(-50%,-50%)', display:'flex', flexDirection:'column', alignItems:'center', gap:3, background:'transparent', border:'none', cursor:'pointer' }}>
            <div style={{ width:34, height:34, borderRadius:'50%', background:`color-mix(in oklch, ${col} 22%, var(--bg-1))`, border:`2px solid ${on?'var(--accent-2)':col}`, display:'grid', placeItems:'center', color:col, boxShadow:on?'0 0 16px -2px rgba(var(--accent-2-rgb),.6)':'none' }}><Icon name="user" size={15} /></div>
            <span style={{ fontFamily:'var(--font-display)', fontSize:10, fontWeight:600, color:'var(--text-2)', whiteSpace:'nowrap' }}>{p.name.split(' ')[0]}</span>
          </button>
        )
      })}
    </div>
  )
}

const SIGNAL_TIMELINE = [
  { c:'var(--success)', t:'hiring_sdrs detected · Northwind', d:'2d ago' },
  { c:'var(--info)', t:'pricing-page visit ×2 · Dana Okafor', d:'1d ago' },
  { c:'#F97316', t:'funding_round · Oakline (Series A)', d:'5h ago' },
  { c:'var(--mint-2)', t:'enrichment refreshed · +1 stakeholder', d:'1h ago' },
]

function GraphPage({ client, onCopilot }) {
  const graph = GRAPHS[client.id]
  const [sel, setSel] = React.useState(graph.prospects[0]?.id ?? null)
  const p = graph.prospects.find(x=>x.id===sel)
  const comp = p && graph.companies.find(c=>c.id===p.company)
  return (
    <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
      <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
        <div style={{ display:'flex', alignItems:'flex-end', justifyContent:'space-between', marginBottom:6 }}>
          <div>
            <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:20, fontWeight:600, color:'var(--text-1)' }}>{client.name} — knowledge graph</h2>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginTop:4, letterSpacing:'.06em' }}>ISOLATED NAMESPACE · {graph.prospects.length} prospects · {graph.companies.length} companies</div>
          </div>
          <Button variant="pixie" size="md" icon="sparkle" onClick={onCopilot}>Enrich with Pixie</Button>
        </div>
        <div style={{ fontSize:12.5, color:'var(--text-3)', marginBottom:14, maxWidth:620 }}>ICP — {client.icp}. This graph is the personalization moat: every open, reply, and call writes back as an edge and recomputes fit + account health.</div>

        <StakeholderMap graph={graph} selected={sel} onSelect={setSel} />

        <div style={{ marginTop:22, fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Prospects · fit + intent</div>
        <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
          {graph.prospects.map(pr => {
            const on = sel===pr.id, col = INTENT_COLOR[pr.intent]
            return (
              <div key={pr.id} onClick={()=>setSel(pr.id)} style={{ display:'flex', alignItems:'center', gap:14, padding:'11px 14px', background:'var(--bg-1)', border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, borderRadius:10, cursor:'pointer' }}>
                <span style={{ width:9, height:9, borderRadius:'50%', background:col, flexShrink:0, boxShadow:`0 0 6px ${col}` }} />
                <div style={{ width:160, flexShrink:0 }}><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13, color:'var(--text-1)' }}>{pr.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>{pr.title}</div></div>
                <div style={{ flex:1, fontSize:12, color:'var(--text-3)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>signals: {pr.signals}</div>
                <div style={{ width:160, flexShrink:0 }}><FitBar fit={pr.fit} intent={pr.intent} /></div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Right: selected prospect + signal timeline + health */}
      <div style={{ width:300, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', overflowY:'auto', padding:'22px 18px' }}>
        {p && (
          <>
            <div style={{ display:'flex', alignItems:'center', gap:12, marginBottom:14 }}>
              <div style={{ width:44, height:44, borderRadius:'50%', background:`color-mix(in oklch, ${INTENT_COLOR[p.intent]} 22%, var(--bg-2))`, border:`2px solid ${INTENT_COLOR[p.intent]}`, display:'grid', placeItems:'center', color:INTENT_COLOR[p.intent] }}><Icon name="user" size={20} /></div>
              <div><div style={{ fontFamily:'var(--font-display)', fontSize:15, fontWeight:600, color:'var(--text-1)' }}>{p.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)' }}>{p.title} · {comp?.name}</div></div>
            </div>
            <div style={{ display:'flex', gap:8, marginBottom:18 }}>
              <div style={{ flex:1, background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, padding:'10px 12px' }}><div style={{ fontFamily:'var(--font-display)', fontWeight:700, fontSize:18, color:INTENT_COLOR[p.intent] }}>{p.fit}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-3)', textTransform:'uppercase', letterSpacing:'.08em' }}>fit score</div></div>
              <div style={{ flex:1, background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, padding:'10px 12px' }}><div style={{ fontFamily:'var(--font-display)', fontWeight:700, fontSize:18, color:'var(--text-1)', textTransform:'capitalize' }}>{p.intent}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-3)', textTransform:'uppercase', letterSpacing:'.08em' }}>intent</div></div>
            </div>
          </>
        )}
        <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>GraphRAG briefing · why-now</div>
        <div style={{ background:'linear-gradient(135deg, rgba(var(--accent-2-rgb),.1), rgba(var(--mint-2-rgb),.05))', border:'1px solid rgba(var(--accent-2-rgb),.22)', borderRadius:12, padding:14, marginBottom:18 }}>
          <div className="t-pixel" style={{ color:'var(--mint-2)', marginBottom:6 }}>PIXIE · INSIGHT</div>
          <div style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.5 }}>{p ? `${p.name.split(' ')[0]} ${p.intent==='hot'?'is a hot opener':'is warming up'} — ${p.title} at ${comp?.name}, fit ${p.fit}. Reason to reach out now: ${p.signals}.` : 'Select a prospect for a grounded briefing.'}</div>
        </div>
        <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Signal timeline</div>
        <div style={{ display:'flex', flexDirection:'column' }}>
          {SIGNAL_TIMELINE.map((s,i)=>(
            <div key={i} style={{ display:'flex', gap:10, padding:'10px 0', borderBottom:i<SIGNAL_TIMELINE.length-1?'1px solid var(--border-1)':'none' }}>
              <span style={{ width:6, height:6, borderRadius:'50%', background:s.c, marginTop:5, flexShrink:0, boxShadow:`0 0 6px ${s.c}` }} />
              <div style={{ flex:1 }}><div style={{ fontSize:12, color:'var(--text-2)', lineHeight:1.45 }}>{s.t}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)', marginTop:2 }}>{s.d}</div></div>
            </div>
          ))}
        </div>
        <div style={{ marginTop:18, padding:'14px', borderRadius:10, background:'var(--bg-2)', border:'1px solid var(--border-1)' }}>
          <div style={{ display:'flex', justifyContent:'space-between', alignItems:'baseline', marginBottom:8 }}><span style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.1em', textTransform:'uppercase', color:'var(--text-3)' }}>Account health</span><span style={{ fontFamily:'var(--font-display)', fontWeight:700, fontSize:16, color:client.accent }}>{client.health}%</span></div>
          <div style={{ height:7, borderRadius:999, background:'var(--bg-3)', overflow:'hidden' }}><div style={{ width:client.health+'%', height:'100%', background:client.accent, borderRadius:999 }} /></div>
        </div>
      </div>
    </div>
  )
}

window.GraphPage = GraphPage

/* ======== Inbox.jsx ======== */
/* ChampIQ (Sim) — Inbox / Replies: classification, human approval gate, KG context. */

function InboxPage({ client, onCopilot, onPersonalize }) {
  const threads = INBOX[client.id] || []
  const [sel, setSel] = React.useState(threads[0]?.id ?? null)
  const [gates, setGates] = React.useState({})
  const t = threads.find(x=>x.id===sel)
  const gateState = t ? (gates[t.id] ?? t.gate) : null

  return (
    <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
      {/* thread list */}
      <div style={{ width:320, flexShrink:0, background:'var(--bg-1)', borderRight:'1px solid var(--border-1)', display:'flex', flexDirection:'column', minHeight:0 }}>
        <div style={{ padding:'14px 16px', borderBottom:'1px solid var(--border-1)', display:'flex', alignItems:'center', gap:8 }}>
          <Icon name="chat" size={15} stroke="var(--accent-1)" />
          <span style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>Inbox</span>
          <Tag color="var(--mint-2)">AI triage</Tag>
          <div style={{ flex:1 }} />
          <span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>{threads.length}</span>
        </div>
        <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto' }}>
          {threads.map(th => {
            const cls = REPLY_CLASS[th.cls], on = sel===th.id
            return (
              <button key={th.id} onClick={()=>setSel(th.id)} style={{ width:'100%', textAlign:'left', display:'flex', flexDirection:'column', gap:5, padding:'12px 16px', borderBottom:'1px solid var(--border-1)', borderLeft:`3px solid ${on?'var(--accent-2)':'transparent'}`, background:on?'var(--bg-2)':'transparent', cursor:'pointer' }}>
                <div style={{ display:'flex', alignItems:'center', gap:8 }}>
                  <span style={{ width:7, height:7, borderRadius:'50%', background:cls.c, flexShrink:0 }} />
                  <span style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13, color:'var(--text-1)', flex:1 }}>{th.from}</span>
                  <span style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)' }}>{th.time}</span>
                </div>
                <div style={{ fontSize:12, color:'var(--text-3)', lineHeight:1.4, overflow:'hidden', textOverflow:'ellipsis', display:'-webkit-box', WebkitLineClamp:2, WebkitBoxOrient:'vertical' }}>{th.snippet}</div>
                <div style={{ display:'flex', gap:6, alignItems:'center' }}><Badge color={cls.c} label={cls.l} />{th.gate==='awaiting' && <Badge color="var(--warn)" label="Needs approval" />}</div>
              </button>
            )
          })}
        </div>
      </div>

      {/* thread reader */}
      <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', background:'var(--bg-0)', display:'flex', flexDirection:'column' }}>
        {t ? (
          <>
            <div style={{ padding:'18px 24px', borderBottom:'1px solid var(--border-1)' }}>
              <div style={{ display:'flex', alignItems:'center', gap:10 }}>
                <span style={{ width:9, height:9, borderRadius:'50%', background:REPLY_CLASS[t.cls].c, boxShadow:`0 0 6px ${REPLY_CLASS[t.cls].c}` }} />
                <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:18, fontWeight:600 }}>{t.from} · {t.co}</h2>
                <Badge color={REPLY_CLASS[t.cls].c} label={REPLY_CLASS[t.cls].l} />
              </div>
              <div style={{ fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-4)', marginTop:6 }}>{t.title} · classified by reply-intent LLM</div>
            </div>
            <div style={{ padding:'20px 24px', flex:1 }}>
              <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, padding:'16px 18px', fontSize:14, color:'var(--text-2)', lineHeight:1.6, maxWidth:620 }}>{t.body}</div>

              {/* approval gate */}
              {t.cls==='positive' && (
                <div style={{ marginTop:18, maxWidth:620, borderRadius:12, border:`1px solid ${gateState==='approved'?'rgba(74,222,128,.4)':'rgba(255,210,63,.4)'}`, background:gateState==='approved'?'rgba(74,222,128,.07)':'rgba(255,210,63,.06)', padding:'16px 18px' }}>
                  <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:8 }}>
                    <Icon name={gateState==='approved'?'check':'user'} size={15} stroke={gateState==='approved'?'var(--success)':'var(--warn)'} />
                    <span style={{ fontFamily:'var(--font-mono)', fontSize:10, letterSpacing:'.08em', textTransform:'uppercase', color:gateState==='approved'?'var(--success)':'var(--warn)' }}>{gateState==='approved'?'approved — call queued':'human-in-the-loop gate'}</span>
                  </div>
                  <div style={{ fontSize:13, color:'var(--text-2)', lineHeight:1.5, marginBottom:12 }}>Positive reply — the Sim run is paused here. Approve to route this hot prospect into a consent-gated ChampVoice call.</div>
                  {gateState!=='approved' ? (
                    <div style={{ display:'flex', gap:8 }}>
                      <Button variant="primary" size="md" icon="voice" onClick={()=>setGates(g=>({...g,[t.id]:'approved'}))}>Approve → ChampVoice</Button>
                      <Button variant="secondary" size="md" icon="mail">Reply by email</Button>
                      <Button variant="ghost" size="md">Snooze</Button>
                    </div>
                  ) : (
                    <div style={{ display:'flex', alignItems:'center', gap:8, fontSize:12.5, color:'var(--text-2)' }}><span style={{ width:6, height:6, borderRadius:'50%', background:'var(--info)', boxShadow:'0 0 6px var(--info)' }} />consent.voice ✓ · DND window checked · call scripted from KG</div>
                  )}
                </div>
              )}
            </div>
          </>
        ) : <div style={{ flex:1, display:'grid', placeItems:'center', color:'var(--text-3)' }}>Select a reply.</div>}
      </div>

      {/* KG context */}
      {t && (
        <div style={{ width:280, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', padding:'20px 18px', overflowY:'auto' }}>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:12 }}>KG context · why-now</div>
          <div style={{ background:'linear-gradient(135deg, rgba(var(--accent-2-rgb),.1), rgba(var(--mint-2-rgb),.05))', border:'1px solid rgba(var(--accent-2-rgb),.22)', borderRadius:12, padding:14, marginBottom:16 }}>
            <div className="t-pixel" style={{ color:'var(--mint-2)', marginBottom:6 }}>GRAPHRAG</div>
            <div style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.5 }}>{t.why}</div>
          </div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Suggested by Pixie</div>
          <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
            <Button variant="pixie" size="md" icon="sparkle" onClick={onPersonalize} style={{ width:'100%', justifyContent:'center' }}>Draft a reply</Button>
            <Button variant="secondary" size="md" icon="graph" style={{ width:'100%', justifyContent:'center' }}>Open in graph</Button>
          </div>
        </div>
      )}
    </div>
  )
}

window.InboxPage = InboxPage

/* ======== Sequence.jsx ======== */
/* ChampIQ (Sim) — ChampMail sequence / cadence builder. */

const STEP_META = {
  email:  { color:'#F97316', icon:'mail',  label:'Email' },
  wait:   { color:'#6b7280', icon:'clock', label:'Wait' },
  branch: { color:'#f59e0b', icon:'branch',label:'Branch' },
  voice:  { color:'#3b82f6', icon:'voice', label:'Voice' },
}

function StepRail({ steps, sel, onSel }) {
  return (
    <div style={{ display:'flex', flexDirection:'column' }}>
      {steps.map((s,i) => {
        const m = STEP_META[s.kind], on = sel===s.id
        return (
          <div key={s.id} style={{ display:'flex', gap:14 }}>
            {/* timeline gutter */}
            <div style={{ display:'flex', flexDirection:'column', alignItems:'center', width:40, flexShrink:0 }}>
              <div style={{ width:34, height:34, borderRadius:9, background:`color-mix(in oklch, ${m.color} 20%, var(--bg-1))`, border:`1.5px solid ${on?'var(--accent-2)':m.color}`, color:m.color, display:'grid', placeItems:'center', boxShadow:on?'0 0 14px -2px rgba(var(--accent-2-rgb),.6)':'none', flexShrink:0 }}><Icon name={m.icon} size={16} /></div>
              {i<steps.length-1 && <div style={{ width:2, flex:1, minHeight:26, background:'var(--border-2)', margin:'2px 0' }} />}
            </div>
            {/* card */}
            <button onClick={()=>onSel(s.id)} style={{ flex:1, textAlign:'left', marginBottom:14, background:on?'var(--bg-2)':'var(--bg-1)', border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, borderRadius:11, padding:'12px 14px', cursor:'pointer', transition:'all .15s' }}>
              <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:s.kind==='wait'||s.kind==='branch'?0:6 }}>
                <span style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.1em', textTransform:'uppercase', color:m.color }}>Day {s.day} · {m.label}</span>
                {s.ai && <Tag color="var(--mint-2)">AI</Tag>}
                {s.gated && <Tag color="var(--info)">consent-gated</Tag>}
                <div style={{ flex:1 }} />
                {s.reply!=null && <span style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)' }}>{s.open}% open · {s.reply}% reply</span>}
              </div>
              <div style={{ fontFamily:'var(--font-display)', fontSize:13.5, fontWeight:600, color:'var(--text-1)' }}>{s.title}</div>
              {s.kind==='email' && <div style={{ fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-3)', marginTop:3 }}>Subj: {s.subject}</div>}
              {(s.kind==='wait'||s.kind==='branch'||s.kind==='voice') && <div style={{ fontSize:11.5, color:'var(--text-3)', marginTop:3, lineHeight:1.4 }}>{s.detail}</div>}
              {s.branches && <div style={{ display:'flex', gap:6, marginTop:8 }}>{s.branches.map(b=><span key={b} style={{ fontFamily:'var(--font-mono)', fontSize:9.5, padding:'2px 7px', borderRadius:5, background:'var(--bg-3)', color:'var(--text-3)' }}>{b}</span>)}</div>}
            </button>
          </div>
        )
      })}
      {/* add step */}
      <div style={{ display:'flex', gap:14 }}>
        <div style={{ width:40, display:'flex', justifyContent:'center', flexShrink:0 }}><div style={{ width:34, height:34, borderRadius:9, border:'1.5px dashed var(--border-2)', display:'grid', placeItems:'center', color:'var(--text-4)' }}><Icon name="plus" size={15} /></div></div>
        <button style={{ flex:1, textAlign:'left', background:'transparent', border:'1px dashed var(--border-2)', borderRadius:11, padding:'12px 14px', cursor:'pointer', color:'var(--text-3)', fontFamily:'var(--font-display)', fontWeight:500, fontSize:13 }}>Add a step — email · wait · branch · voice</button>
      </div>
    </div>
  )
}

function StepInspector({ step, onPersonalize }) {
  if (!step) return null
  const m = STEP_META[step.kind]
  return (
    <div style={{ width:320, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', overflowY:'auto', padding:'18px 18px' }}>
      <div style={{ display:'flex', alignItems:'center', gap:10, marginBottom:16 }}>
        <div style={{ width:32, height:32, borderRadius:8, background:`color-mix(in oklch, ${m.color} 20%, transparent)`, color:m.color, display:'grid', placeItems:'center' }}><Icon name={m.icon} size={16} /></div>
        <div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)' }}>{m.label} · day {step.day}</div><div style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>{step.title}</div></div>
      </div>

      {step.kind==='email' && (
        <div style={{ display:'flex', flexDirection:'column', gap:14 }}>
          <div><label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:4 }}>Subject</label><Input value={step.subject} readOnly /></div>
          <div>
            <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:6 }}>
              <label style={{ fontSize:11, color:'var(--text-3)' }}>AI personalization</label>
              <Tag color={step.ai?'var(--mint-2)':'var(--text-4)'}>{step.ai?'On · GraphRAG':'Off · static'}</Tag>
            </div>
            <div style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.5, padding:'10px 12px', background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8 }}>{step.preview}</div>
          </div>
          {step.ai && <Button variant="pixie" size="md" icon="sparkle" onClick={onPersonalize} style={{ width:'100%', justifyContent:'center' }}>Preview personalization</Button>}
          <div style={{ display:'flex', gap:10 }}>
            <div style={{ flex:1, background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, padding:'10px 12px' }}><div style={{ fontFamily:'var(--font-display)', fontWeight:700, fontSize:18, color:'#F97316' }}>{step.open}%</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-3)', textTransform:'uppercase', letterSpacing:'.08em' }}>open</div></div>
            <div style={{ flex:1, background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, padding:'10px 12px' }}><div style={{ fontFamily:'var(--font-display)', fontWeight:700, fontSize:18, color:'var(--success)' }}>{step.reply}%</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-3)', textTransform:'uppercase', letterSpacing:'.08em' }}>reply</div></div>
          </div>
        </div>
      )}
      {step.kind==='wait' && (
        <div><label style={{ fontSize:11, color:'var(--text-3)', display:'block', marginBottom:4 }}>Delay</label><Input value={step.title.replace('Wait ','')} readOnly /><div style={{ fontSize:12, color:'var(--text-3)', marginTop:10, lineHeight:1.5 }}>{step.detail}</div></div>
      )}
      {step.kind==='branch' && (
        <div><div style={{ fontSize:12, color:'var(--text-3)', lineHeight:1.5, marginBottom:10 }}>{step.detail}</div>{step.branches.map(b=><div key={b} style={{ display:'flex', alignItems:'center', gap:8, padding:'9px 11px', background:'var(--bg-2)', border:'1px solid var(--border-1)', borderRadius:8, marginBottom:6, fontSize:12, color:'var(--text-2)' }}><Icon name="branch" size={13} stroke="var(--warn)" />{b}</div>)}</div>
      )}
      {step.kind==='voice' && (
        <div style={{ display:'flex', alignItems:'center', gap:8, padding:'12px 13px', borderRadius:10, background:'rgba(91,192,255,.08)', border:'1px solid rgba(91,192,255,.3)' }}><Icon name="check" size={15} stroke="var(--info)" /><span style={{ fontSize:12.5, color:'var(--text-2)', lineHeight:1.5 }}>{step.detail}</span></div>
      )}
    </div>
  )
}

function SequencePage({ onPersonalize }) {
  const seq = SEQUENCE
  const [sel, setSel] = React.useState(seq.steps[0].id)
  const step = seq.steps.find(s=>s.id===sel)
  const emails = seq.steps.filter(s=>s.kind==='email').length
  return (
    <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
      <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
        <div style={{ display:'flex', alignItems:'flex-end', justifyContent:'space-between', marginBottom:18, flexWrap:'wrap', gap:12 }}>
          <div>
            <div style={{ display:'flex', alignItems:'center', gap:10 }}>
              <Icon name="mail" size={18} stroke="var(--accent-1)" />
              <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:20, fontWeight:600 }}>{seq.name}</h2>
              <Tag color="var(--mint-2)">ChampMail cadence</Tag>
            </div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginTop:6, letterSpacing:'.04em' }}>{seq.steps.length} STEPS · {emails} EMAILS · {seq.enrolled} ENROLLED · {seq.sendWindow}</div>
          </div>
          <div style={{ display:'flex', gap:8 }}>
            <Button variant="ghost" size="md" icon="eye" onClick={onPersonalize}>Preview</Button>
            <Button variant="secondary" size="md">Pause</Button>
            <Button variant="primary" size="md" icon="play">Active</Button>
          </div>
        </div>

        {/* enrolment funnel mini */}
        <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12, marginBottom:24 }}>
          {[['84','Enrolled','var(--accent-2)'],['62%','Avg open','#F97316'],['11%','Avg reply','var(--success)'],['9','Meetings','var(--info)']].map(([v,l,c])=>(
            <div key={l} style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'13px 16px' }}><div style={{ fontFamily:'var(--font-display)', fontSize:20, fontWeight:700, color:c }}>{v}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.08em', textTransform:'uppercase', color:'var(--text-3)', marginTop:2 }}>{l}</div></div>
          ))}
        </div>

        <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:14 }}>Cadence</div>
        <StepRail steps={seq.steps} sel={sel} onSel={setSel} />
      </div>
      <StepInspector step={step} onPersonalize={onPersonalize} />
    </div>
  )
}

window.SequencePage = SequencePage

/* ======== Personalize.jsx ======== */
/* ChampIQ (Sim) — per-prospect AI personalization preview (GraphRAG grounding). */

// Render a segment array; grounded spans get an accent underline + hover → fact.
function GroundedText({ parts, activeFact, onHover, big }) {
  return (
    <span style={{ whiteSpace:'pre-wrap', lineHeight:big?1.3:1.65, fontSize:big?19:14.5, fontWeight:big?600:400, color:'var(--text-1)', fontFamily:big?'var(--font-display)':'var(--font-body)', letterSpacing:big?'-.01em':'normal' }}>
      {parts.map((seg,i) => seg.g ? (
        <mark key={i}
          onMouseEnter={()=>onHover&&onHover(seg.g)} onMouseLeave={()=>onHover&&onHover(null)}
          style={{ background: activeFact===seg.g ? 'rgba(var(--accent-2-rgb),.32)' : 'rgba(var(--accent-2-rgb),.14)', color:'var(--text-1)', borderBottom:'2px solid var(--accent-2)', borderRadius:'3px 3px 0 0', padding:'0 2px', cursor:'help', transition:'background .15s' }}>
          {seg.t}
        </mark>
      ) : <React.Fragment key={i}>{seg.t}</React.Fragment>)}
    </span>
  )
}

function PersonalizePage({ client, onOpenInbox }) {
  const list = (PROSPECTS[client.id] || []).filter(p=>!p.suppressed)
  const [pid, setPid] = React.useState(list[0]?.id ?? null)
  const [mode, setMode] = React.useState('email') // email | voice
  const [activeFact, setActiveFact] = React.useState(null)
  const [showAlt, setShowAlt] = React.useState(false)
  const data = PERSONALIZE[pid] || PERSONALIZE_FALLBACK
  const prospect = list.find(p=>p.id===pid)
  const groundedCount = [...data.subject, ...data.body].filter(s=>s.g).length

  return (
    <div style={{ flex:1, display:'flex', minHeight:0, minWidth:0, overflow:'hidden' }}>
      {/* prospect picker */}
      <div style={{ width:240, flexShrink:0, background:'var(--bg-1)', borderRight:'1px solid var(--border-1)', display:'flex', flexDirection:'column', minHeight:0 }}>
        <div style={{ padding:'14px 16px', borderBottom:'1px solid var(--border-1)', fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)' }}>Prospects</div>
        <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto' }}>
          {list.map(p => {
            const on = pid===p.id, col = INTENT_COLOR[p.intent]
            return (
              <button key={p.id} onClick={()=>{ setPid(p.id); setActiveFact(null) }} style={{ width:'100%', textAlign:'left', display:'flex', alignItems:'center', gap:10, padding:'11px 16px', borderBottom:'1px solid var(--border-1)', borderLeft:`3px solid ${on?'var(--accent-2)':'transparent'}`, background:on?'var(--bg-2)':'transparent', cursor:'pointer' }}>
                <div style={{ width:28, height:28, borderRadius:'50%', background:`color-mix(in oklch, ${col} 20%, var(--bg-2))`, border:`1.5px solid ${col}`, display:'grid', placeItems:'center', color:col, flexShrink:0, fontFamily:'var(--font-display)', fontWeight:600, fontSize:11 }}>{p.name[0]}</div>
                <div style={{ flex:1, minWidth:0 }}><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:12.5, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{p.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)' }}>{p.company} · fit {p.fit}</div></div>
              </button>
            )
          })}
        </div>
      </div>

      {/* generated message */}
      <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', background:'var(--bg-0)', padding:'24px 28px' }}>
        <div style={{ display:'flex', alignItems:'flex-end', justifyContent:'space-between', marginBottom:6, gap:12, flexWrap:'wrap' }}>
          <div>
            <div style={{ display:'flex', alignItems:'center', gap:10 }}>
              <Icon name="sparkle" size={17} stroke="var(--mint-2)" />
              <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:19, fontWeight:600 }}>Personalization preview</h2>
            </div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginTop:5, letterSpacing:'.04em' }}>{prospect?.name.toUpperCase()} · GROUNDED IN {groundedCount} KG FACTS · NOT A TEMPLATE</div>
          </div>
          <div style={{ display:'flex', gap:6 }}>
            {[['email','Email','mail'],['voice','Voice script','voice']].map(([k,l,ic])=>{
              const on=mode===k
              return <button key={k} onClick={()=>setMode(k)} style={{ display:'flex', alignItems:'center', gap:6, padding:'6px 12px', borderRadius:8, cursor:'pointer', fontFamily:'var(--font-display)', fontWeight:600, fontSize:12.5, border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, background:on?'rgba(var(--accent-2-rgb),.14)':'var(--bg-1)', color:on?'var(--accent-1)':'var(--text-3)' }}><Icon name={ic} size={13} />{l}</button>
            })}
          </div>
        </div>

        {/* email composer card */}
        <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:14, overflow:'hidden', marginTop:18, maxWidth:680 }}>
          <div style={{ display:'flex', alignItems:'center', gap:8, padding:'12px 18px', borderBottom:'1px solid var(--border-1)', background:'var(--bg-2)' }}>
            <span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)', textTransform:'uppercase', letterSpacing:'.08em' }}>To</span>
            <span style={{ fontSize:12.5, color:'var(--text-2)' }}>{prospect?.name} · {prospect?.company}</span>
            <div style={{ flex:1 }} />
            {mode==='email' && <button onClick={()=>setShowAlt(a=>!a)} style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--accent-1)', background:'rgba(var(--accent-2-rgb),.1)', border:'1px solid rgba(var(--accent-2-rgb),.25)', borderRadius:999, padding:'3px 10px', cursor:'pointer' }}>{showAlt?'Variant B':'Variant A'} · A/B</button>}
          </div>
          <div style={{ padding:'18px 22px' }}>
            {mode==='email' ? (
              <>
                <div style={{ paddingBottom:14, borderBottom:'1px solid var(--border-1)', marginBottom:14 }}>
                  <div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)', textTransform:'uppercase', letterSpacing:'.1em', marginBottom:6 }}>Subject</div>
                  {showAlt ? <span style={{ fontFamily:'var(--font-display)', fontSize:19, fontWeight:600, color:'var(--text-1)' }}>{data.subjectAlt}</span> : <GroundedText parts={data.subject} activeFact={activeFact} onHover={setActiveFact} big />}
                </div>
                <GroundedText parts={data.body} activeFact={activeFact} onHover={setActiveFact} />
              </>
            ) : (
              <>
                <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:12 }}><Icon name="voice" size={15} stroke="var(--info)" /><span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-3)', textTransform:'uppercase', letterSpacing:'.08em' }}>ChampVoice script · ElevenLabs dynamic vars</span></div>
                <GroundedText parts={data.voice} activeFact={activeFact} onHover={setActiveFact} />
                <div style={{ display:'flex', alignItems:'center', gap:8, marginTop:16, padding:'10px 12px', borderRadius:8, background:'rgba(91,192,255,.08)', border:'1px solid rgba(91,192,255,.3)' }}><Icon name="check" size={13} stroke="var(--info)" /><span style={{ fontSize:12, color:'var(--text-2)' }}>Fires only on hot intent + voice consent · DND window checked.</span></div>
              </>
            )}
          </div>
          <div style={{ display:'flex', gap:8, padding:'14px 18px', borderTop:'1px solid var(--border-1)', background:'var(--bg-2)' }}>
            <Button variant="primary" size="md" icon={mode==='email'?'send':'voice'}>{mode==='email'?'Approve & send':'Queue call (gated)'}</Button>
            <Button variant="pixie" size="md" icon="sparkle">Regenerate</Button>
            <div style={{ flex:1 }} />
            <Button variant="ghost" size="md" icon="chat" onClick={onOpenInbox}>Open thread</Button>
          </div>
        </div>

        <div style={{ display:'flex', alignItems:'center', gap:8, marginTop:14, fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-4)' }}><span style={{ width:14, height:8, borderRadius:2, background:'rgba(var(--accent-2-rgb),.3)', borderBottom:'2px solid var(--accent-2)' }} />Highlighted = grounded in a live KG fact. Hover to trace it.</div>
      </div>

      {/* grounding panel */}
      <div style={{ width:288, flexShrink:0, background:'var(--bg-1)', borderLeft:'1px solid var(--border-1)', overflowY:'auto', padding:'20px 18px' }}>
        <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:12 }}>GraphRAG grounding</div>
        <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
          {data.facts.map(f => {
            const on = activeFact===f.id
            return (
              <div key={f.id} onMouseEnter={()=>setActiveFact(f.id)} onMouseLeave={()=>setActiveFact(null)} style={{ padding:'11px 13px', borderRadius:10, border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, background:on?'rgba(var(--accent-2-rgb),.08)':'var(--bg-2)', cursor:'help', transition:'all .15s' }}>
                <div style={{ display:'flex', alignItems:'center', gap:7, marginBottom:4 }}>
                  <span style={{ width:7, height:7, borderRadius:'50%', background:'var(--accent-2)', flexShrink:0 }} />
                  <span style={{ fontFamily:'var(--font-display)', fontSize:12.5, fontWeight:600, color:'var(--text-1)' }}>{f.label}</span>
                </div>
                <div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)', paddingLeft:14 }}>{f.source}</div>
                <div style={{ display:'flex', alignItems:'center', gap:6, paddingLeft:14, marginTop:6 }}>
                  <div style={{ flex:1, height:4, borderRadius:999, background:'var(--bg-3)', overflow:'hidden' }}><div style={{ width:`${Math.round(f.conf*100)}%`, height:'100%', background:f.conf>0.7?'var(--success)':'var(--warn)' }} /></div>
                  <span style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-4)' }}>{Math.round(f.conf*100)}%</span>
                </div>
              </div>
            )
          })}
        </div>
        <div style={{ marginTop:16, padding:'12px 13px', borderRadius:10, background:'linear-gradient(135deg, rgba(var(--accent-2-rgb),.1), rgba(var(--mint-2-rgb),.05))', border:'1px solid rgba(var(--accent-2-rgb),.22)' }}>
          <div className="t-pixel" style={{ color:'var(--mint-2)', marginBottom:6 }}>PIXIE · NOTE</div>
          <div style={{ fontSize:12, color:'var(--text-2)', lineHeight:1.5 }}>Insight over content — I write from facts the graph can prove, never invented claims. Low-confidence facts are left out.</div>
        </div>
      </div>
    </div>
  )
}

window.PersonalizePage = PersonalizePage

/* ======== Prospects.jsx ======== */
/* ChampIQ (Sim) — Prospects table: search, fit/intent, suppression, bulk enroll. */

const STAGE_COLOR = { Replied:'var(--success)', Sequenced:'var(--accent-2)', Enriched:'var(--info)', New:'var(--text-3)', Suppressed:'var(--text-4)' }

function ProspectsPage({ client, onOpenGraph, onEnroll }) {
  const all = PROSPECTS[client.id] || []
  const [q, setQ] = React.useState('')
  const [intent, setIntent] = React.useState('all')
  const [picked, setPicked] = React.useState({})
  const rows = all.filter(p =>
    (intent==='all' || p.intent===intent) &&
    (q==='' || (p.name+p.company+p.title).toLowerCase().includes(q.toLowerCase()))
  )
  const pickedIds = Object.keys(picked).filter(k=>picked[k])
  const allPicked = rows.length>0 && rows.every(r=>picked[r.id])

  return (
    <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
      <div style={{ display:'flex', alignItems:'flex-end', justifyContent:'space-between', marginBottom:16, gap:12, flexWrap:'wrap' }}>
        <div>
          <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:20, fontWeight:600 }}>Prospects</h2>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginTop:4, letterSpacing:'.06em' }}>{all.length} IN {client.name.toUpperCase()} · client:{client.id}</div>
        </div>
        <div style={{ display:'flex', gap:8, alignItems:'center' }}>
          <div style={{ width:240 }}><Input placeholder="Search prospects…" icon="search" value={q} onChange={e=>setQ(e.target.value)} /></div>
          <Button variant="secondary" size="md" icon="plus">Import CSV</Button>
        </div>
      </div>

      {/* filter chips */}
      <div style={{ display:'flex', gap:6, marginBottom:14 }}>
        {[['all','All'],['hot','Hot'],['warm','Warm'],['cold','Cold']].map(([k,l])=>{
          const on = intent===k
          return <button key={k} onClick={()=>setIntent(k)} style={{ padding:'5px 12px', borderRadius:999, fontFamily:'var(--font-mono)', fontSize:10.5, letterSpacing:'.06em', textTransform:'uppercase', cursor:'pointer', border:`1px solid ${on?'var(--accent-2)':'var(--border-1)'}`, background:on?'rgba(var(--accent-2-rgb),.14)':'var(--bg-1)', color:on?'var(--accent-1)':'var(--text-3)' }}>{l}</button>
        })}
      </div>

      {/* table */}
      <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, overflow:'hidden' }}>
        <div style={{ display:'grid', gridTemplateColumns:'34px 1.6fr 1fr 150px 1fr 110px', alignItems:'center', gap:12, padding:'10px 16px', borderBottom:'1px solid var(--border-1)', background:'var(--bg-2)', fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.12em', textTransform:'uppercase', color:'var(--text-4)' }}>
          <button onClick={()=>{ const np={}; if(!allPicked) rows.forEach(r=>np[r.id]=true); setPicked(np) }} style={{ width:16, height:16, borderRadius:4, border:`1.5px solid ${allPicked?'var(--accent-2)':'var(--border-2)'}`, background:allPicked?'var(--accent-2)':'transparent', cursor:'pointer', display:'grid', placeItems:'center' }}>{allPicked&&<Icon name="check" size={11} stroke="#fff" />}</button>
          <span>Prospect</span><span>Company</span><span>Fit · intent</span><span>Stage / sequence</span><span style={{ textAlign:'right' }}>Actions</span>
        </div>
        {rows.map(p => {
          const on = !!picked[p.id], col = INTENT_COLOR[p.intent]
          return (
            <div key={p.id} style={{ display:'grid', gridTemplateColumns:'34px 1.6fr 1fr 150px 1fr 110px', alignItems:'center', gap:12, padding:'12px 16px', borderBottom:'1px solid var(--border-1)', opacity:p.suppressed?0.6:1 }}>
              <button onClick={()=>setPicked(s=>({...s,[p.id]:!s[p.id]}))} style={{ width:16, height:16, borderRadius:4, border:`1.5px solid ${on?'var(--accent-2)':'var(--border-2)'}`, background:on?'var(--accent-2)':'transparent', cursor:'pointer', display:'grid', placeItems:'center' }}>{on&&<Icon name="check" size={11} stroke="#fff" />}</button>
              <div style={{ display:'flex', alignItems:'center', gap:10 }}>
                <div style={{ width:30, height:30, borderRadius:'50%', background:`color-mix(in oklch, ${col} 20%, var(--bg-2))`, border:`1.5px solid ${col}`, display:'grid', placeItems:'center', color:col, flexShrink:0, fontFamily:'var(--font-display)', fontWeight:600, fontSize:12 }}>{p.name[0]}</div>
                <div><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13 }}>{p.name}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>{p.title}</div></div>
              </div>
              <div style={{ fontSize:12.5, color:'var(--text-2)' }}>{p.company}</div>
              <div style={{ display:'flex', alignItems:'center', gap:8 }}>
                <div style={{ flex:1, height:6, borderRadius:999, background:'var(--bg-3)', overflow:'hidden' }}><div style={{ width:p.fit+'%', height:'100%', background:col }} /></div>
                <span style={{ fontFamily:'var(--font-mono)', fontSize:10, color:col, width:30 }}>{p.fit}</span>
              </div>
              <div style={{ display:'flex', flexDirection:'column', gap:2 }}>
                <Badge color={STAGE_COLOR[p.stage]} label={p.stage} />
                {p.seq!=='—' && <span style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)' }}>{p.seq}</span>}
              </div>
              <div style={{ display:'flex', justifyContent:'flex-end', gap:6 }}>
                <button onClick={onOpenGraph} title="View in graph" style={{ width:28, height:28, display:'grid', placeItems:'center', borderRadius:7, border:'1px solid var(--border-1)', background:'var(--bg-2)', color:'var(--text-3)', cursor:'pointer' }}><Icon name="graph" size={14} /></button>
                <button title="Enrich" style={{ width:28, height:28, display:'grid', placeItems:'center', borderRadius:7, border:'1px solid var(--border-1)', background:'var(--bg-2)', color:'var(--text-3)', cursor:'pointer' }}><Icon name="refresh" size={14} /></button>
              </div>
            </div>
          )
        })}
        {rows.length===0 && <div style={{ padding:'40px', textAlign:'center', color:'var(--text-3)', fontSize:13 }}>No prospects match.</div>}
      </div>

      {/* bulk action bar */}
      {pickedIds.length>0 && (
        <div className="ciq-fade-in" style={{ position:'sticky', bottom:16, marginTop:16, display:'flex', alignItems:'center', gap:14, padding:'12px 18px', background:'var(--bg-2)', border:'1px solid var(--accent-2)', borderRadius:12, boxShadow:'var(--shadow-lg)' }}>
          <span style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13 }}>{pickedIds.length} selected</span>
          <div style={{ flex:1 }} />
          <Button variant="ghost" size="md" icon="refresh">Re-enrich</Button>
          <Button variant="secondary" size="md" icon="archive">Suppress</Button>
          <Button variant="primary" size="md" icon="play" onClick={onEnroll}>Enroll in sequence</Button>
        </div>
      )}
    </div>
  )
}

window.ProspectsPage = ProspectsPage

/* ======== Deliverability.jsx ======== */
/* ChampIQ (Sim) — Deliverability: per-client sending domains, warmup, suppression. */

function DomainCard({ d }) {
  const warming = d.state==='warming'
  const col = warming ? 'var(--warn)' : 'var(--success)'
  const pct = Math.round(d.sent / d.cap * 100)
  return (
    <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, padding:16 }}>
      <div style={{ display:'flex', alignItems:'center', gap:10, marginBottom:12 }}>
        <div style={{ width:34, height:34, borderRadius:9, background:'color-mix(in oklch, var(--accent-2) 18%, transparent)', color:'var(--accent-1)', display:'grid', placeItems:'center' }}><Icon name="mail" size={16} /></div>
        <div style={{ flex:1 }}>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:12.5, color:'var(--text-1)' }}>{d.domain}</div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>reputation {d.rep}/100</div>
        </div>
        <Badge color={col} label={warming?`Warmup · day ${d.day}`:'Healthy'} />
      </div>
      <div style={{ display:'flex', justifyContent:'space-between', fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginBottom:5 }}><span>Today {d.sent}/{d.cap}</span><span>{pct}% of cap</span></div>
      <div style={{ height:7, borderRadius:999, background:'var(--bg-3)', overflow:'hidden' }}><div style={{ width:pct+'%', height:'100%', background:col, borderRadius:999 }} /></div>
      {/* warmup ramp sparkline */}
      <div style={{ display:'flex', alignItems:'flex-end', gap:3, height:34, marginTop:14 }}>
        {Array.from({length:14}).map((_,i)=>{
          const h = warming ? Math.min(100, 18 + i*6) : 100
          const active = i <= (warming ? d.day-1 : 13)
          return <div key={i} style={{ flex:1, height:`${h}%`, borderRadius:2, background:active?col:'var(--bg-3)', opacity:active?1:.5 }} />
        })}
      </div>
      <div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-4)', marginTop:6, letterSpacing:'.06em' }}>30-DAY WARMUP RAMP</div>
    </div>
  )
}

function DeliverabilityPage({ client }) {
  const domains = DOMAINS[client.id] || []
  const metrics = [
    ['0.04%','Spam rate','var(--success)','< 0.1% target'],
    ['0.6%','Bounce rate','var(--info)','< 2% target'],
    ['98.9%','Inbox placement','var(--accent-2)','last 7d'],
    [`${SUPPRESSION.length}`,'Suppressed','var(--text-2)','do-not-contact'],
  ]
  return (
    <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
      <div style={{ marginBottom:6 }}>
        <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:20, fontWeight:600 }}>Deliverability</h2>
        <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginTop:4, letterSpacing:'.06em' }}>OWNED CHANNEL · {client.name.toUpperCase()} · PER-CLIENT DOMAIN ISOLATION</div>
      </div>
      <div style={{ fontSize:12.5, color:'var(--text-3)', marginBottom:20, maxWidth:640 }}>Reputation is managed per client, not on a shared ESP. Every send passes suppression → consent → warmup cap → the client's own sending domain.</div>

      <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12, marginBottom:24 }}>
        {metrics.map(([v,l,c,n])=>(
          <div key={l} style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'14px 16px' }}>
            <div style={{ fontFamily:'var(--font-display)', fontSize:22, fontWeight:700, color:c, letterSpacing:'-.02em' }}>{v}</div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.08em', textTransform:'uppercase', color:'var(--text-3)', marginTop:2 }}>{l}</div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:9, color:'var(--text-4)', marginTop:2 }}>{n}</div>
          </div>
        ))}
      </div>

      <div style={{ display:'grid', gridTemplateColumns:'1.5fr 1fr', gap:18 }}>
        <div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10, display:'flex', justifyContent:'space-between' }}><span>Sending domains</span><span style={{ color:'var(--text-3)' }}>Stalwart MTA</span></div>
          <div style={{ display:'grid', gridTemplateColumns:domains.length>1?'1fr 1fr':'1fr', gap:12 }}>
            {domains.map(d=><DomainCard key={d.domain} d={d} />)}
          </div>
        </div>
        <div>
          <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Suppression list</div>
          <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, overflow:'hidden' }}>
            {SUPPRESSION.map((s,i)=>(
              <div key={i} style={{ display:'flex', alignItems:'center', gap:10, padding:'11px 14px', borderBottom:i<SUPPRESSION.length-1?'1px solid var(--border-1)':'none' }}>
                <Icon name="archive" size={14} stroke="var(--text-4)" />
                <div style={{ flex:1, minWidth:0 }}><div style={{ fontFamily:'var(--font-mono)', fontSize:11.5, color:'var(--text-2)', overflow:'hidden', textOverflow:'ellipsis' }}>{s.addr}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)' }}>{s.date}</div></div>
                <Tag color="var(--text-4)">{s.reason}</Tag>
              </div>
            ))}
          </div>
          <div style={{ display:'flex', alignItems:'center', gap:8, marginTop:12, padding:'11px 13px', borderRadius:10, background:'rgba(91,192,255,.08)', border:'1px solid rgba(91,192,255,.3)' }}>
            <Icon name="check" size={14} stroke="var(--info)" /><span style={{ fontSize:12, color:'var(--text-2)' }}>List-Unsubscribe + one-click headers on every send.</span>
          </div>
        </div>
      </div>
    </div>
  )
}

window.DeliverabilityPage = DeliverabilityPage

/* ======== MailAnalytics.jsx ======== */
/* ChampIQ (Sim) — ChampMail engine + client Analytics pages. */

const MAIL_ROWS = [
  { from:'Dana Okafor', co:'Northwind', subj:'Re: saw you\u2019re scaling the sales team', intent:'positive', time:'2m' },
  { from:'Sam Cole', co:'Oakline', subj:'Re: congrats on the raise', intent:'positive', time:'40m' },
  { from:'Marco Reyes', co:'Lumen Labs', subj:'Not the right time', intent:'negative', time:'1h' },
  { from:'Pat Reyes', co:'Deadlink', subj:'unsubscribe', intent:'unsub', time:'3h' },
]
const INTENT_PILL = { positive:{c:'var(--success)',l:'Positive'}, negative:{c:'var(--danger)',l:'Not now'}, unsub:{c:'var(--text-4)',l:'Unsub → suppressed'}, question:{c:'var(--info)',l:'Question'} }

function MailPage({ client }) {
  const deliver = [
    { label:'Suppression gate', val:'on', good:true, note:'checked before every send' },
    { label:'Warmup ramp', val:'day 12 · 50/day', good:true, note:'per-mailbox, grows to cap' },
    { label:'Sending domains', val:`${client.name.split(' ')[0].toLowerCase()}-mail.io · +2`, good:true, note:'per-client isolation' },
    { label:'Spam rate', val:'0.04%', good:true, note:'< 0.1% target' },
  ]
  return (
    <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', background:'var(--bg-0)' }}>
      <div style={{ padding:'24px 28px' }}>
        <div style={{ display:'flex', alignItems:'center', gap:10, marginBottom:4 }}>
          <Icon name="mail" size={18} stroke="var(--accent-1)" />
          <h2 style={{ margin:0, fontFamily:'var(--font-display)', fontSize:20, fontWeight:600, color:'var(--text-1)' }}>ChampMail</h2>
          <Tag color="var(--mint-2)">AI triage</Tag>
          <Tag color="var(--text-4)">Stalwart MTA · SMTP + IMAP</Tag>
        </div>
        <div style={{ fontSize:12.5, color:'var(--text-3)', marginBottom:18 }}>Own the channel — self-hosted MTA, KG-grounded copy, deliverability discipline in the send path. Drives ChampMail engine via Sim blocks; inbound replies become canvas signals.</div>

        <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12, marginBottom:18 }}>
          {[['2,418','Sent · 7d','var(--accent-2)'],['41%','Open rate','#F97316'],[`${client.replies}`,'Warm replies','var(--success)'],['1.2%','Reply rate','var(--info)']].map(([v,l,c])=>(
            <div key={l} style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'14px 16px' }}><div style={{ fontFamily:'var(--font-display)', fontSize:22, fontWeight:700, color:c }}>{v}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.1em', textTransform:'uppercase', color:'var(--text-3)', marginTop:2 }}>{l}</div></div>
          ))}
        </div>

        <div style={{ display:'grid', gridTemplateColumns:'1.6fr 1fr', gap:18 }}>
          {/* Inbox */}
          <div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Triaged inbox</div>
            {MAIL_ROWS.map((r,i)=>{ const it=INTENT_PILL[r.intent]; return (
              <div key={i} style={{ display:'flex', alignItems:'center', gap:14, padding:'12px 14px', background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, marginBottom:8 }}>
                <div style={{ width:34, height:34, borderRadius:'50%', background:'var(--bg-3)', display:'grid', placeItems:'center', color:'var(--text-2)', fontFamily:'var(--font-display)', fontWeight:600, fontSize:13 }}>{r.from[0]}</div>
                <div style={{ width:130, flexShrink:0 }}><div style={{ fontFamily:'var(--font-display)', fontWeight:600, fontSize:13, color:'var(--text-1)' }}>{r.from}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:10, color:'var(--text-4)' }}>{r.co}</div></div>
                <div style={{ flex:1, fontSize:13, color:'var(--text-2)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{r.subj}</div>
                <Badge color={it.c} label={it.l} />
              </div>
            )})}
          </div>
          {/* Deliverability */}
          <div>
            <div style={{ fontFamily:'var(--font-mono)', fontSize:9, letterSpacing:'.14em', textTransform:'uppercase', color:'var(--text-4)', marginBottom:10 }}>Deliverability</div>
            <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
              {deliver.map(d=>(
                <div key={d.label} style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'11px 13px' }}>
                  <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:2 }}>
                    <span style={{ display:'flex', alignItems:'center', gap:6, fontSize:12.5, color:'var(--text-2)' }}><Icon name="check" size={13} stroke="var(--success)" />{d.label}</span>
                    <span style={{ fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-1)' }}>{d.val}</span>
                  </div>
                  <div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-4)', paddingLeft:19 }}>{d.note}</div>
                </div>
              ))}
              <div style={{ display:'flex', alignItems:'center', gap:8, padding:'11px 13px', borderRadius:10, background:'rgba(91,192,255,.08)', border:'1px solid rgba(91,192,255,.3)' }}>
                <Icon name="check" size={14} stroke="var(--info)" /><span style={{ fontSize:12, color:'var(--text-2)' }}>One-click unsubscribe + List-Unsubscribe headers on every send.</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function Funnel({ rows }) {
  const max = rows[0].v
  return (
    <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
      {rows.map(r=>(
        <div key={r.label} style={{ display:'flex', alignItems:'center', gap:12 }}>
          <span style={{ width:150, flexShrink:0, fontSize:12.5, color:'var(--text-2)' }}>{r.label}</span>
          <div style={{ flex:1, height:26, borderRadius:7, background:'var(--bg-2)', border:'1px solid var(--border-1)', overflow:'hidden', position:'relative' }}>
            <div style={{ width:(r.v/max*100)+'%', height:'100%', background:`linear-gradient(90deg, color-mix(in oklch, ${r.c} 70%, transparent), ${r.c})`, borderRadius:6 }} />
            <span style={{ position:'absolute', left:10, top:'50%', transform:'translateY(-50%)', fontFamily:'var(--font-mono)', fontSize:11, color:'#fff', fontWeight:600 }}>{r.v.toLocaleString()}</span>
          </div>
          <span style={{ width:48, textAlign:'right', fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-3)' }}>{r.pct}</span>
        </div>
      ))}
    </div>
  )
}

function AnalyticsPage({ client }) {
  const funnel = [
    { label:'Signals detected', v:client.signals*10, pct:'100%', c:'#F97316' },
    { label:'Prospects enriched', v:client.prospects, pct:'—', c:'var(--accent-2)' },
    { label:'Emails sent', v:2418, pct:'—', c:'var(--info)' },
    { label:'Replies', v:client.replies*8, pct:'1.2%', c:'var(--mint-2)' },
    { label:'Meetings booked', v:client.replies, pct:'0.4%', c:'var(--success)' },
  ]
  return (
    <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
      <h2 style={{ margin:'0 0 4px', fontFamily:'var(--font-display)', fontSize:20, fontWeight:600, color:'var(--text-1)' }}>{client.name} — analytics</h2>
      <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginBottom:20, letterSpacing:'.06em' }}>SIGNAL → MEETING FUNNEL · LAST 30 DAYS</div>

      <div style={{ display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12, marginBottom:24 }}>
        {[['18–25%','Reply rate (signal-first)','var(--success)'],['3–5×','Lift vs templates','var(--accent-2)'],['0.04%','Spam rate','var(--info)'],[`${client.health}%`,'Account health','#F97316']].map(([v,l,c])=>(
          <div key={l} style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:10, padding:'16px 18px' }}><div style={{ fontFamily:'var(--font-display)', fontSize:24, fontWeight:700, color:c, letterSpacing:'-.02em' }}>{v}</div><div style={{ fontFamily:'var(--font-mono)', fontSize:9.5, color:'var(--text-3)', marginTop:3, textTransform:'uppercase', letterSpacing:'.08em' }}>{l}</div></div>
        ))}
      </div>

      <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, padding:20, marginBottom:18 }}>
        <div style={{ fontFamily:'var(--font-display)', fontSize:15, fontWeight:600, color:'var(--text-1)', marginBottom:14 }}>Pipeline funnel</div>
        <Funnel rows={funnel} />
      </div>

      <div style={{ display:'flex', gap:14, alignItems:'center', padding:'16px 18px', borderRadius:12, background:'linear-gradient(135deg, rgba(var(--accent-2-rgb),.1), rgba(var(--mint-2-rgb),.05))', border:'1px solid rgba(var(--accent-2-rgb),.22)' }}>
        <Pixie pose="read" size={72} />
        <div style={{ flex:1 }}>
          <div className="t-pixel" style={{ color:'var(--mint-2)', marginBottom:4 }}>PIXIE · INSIGHT</div>
          <div style={{ fontSize:13.5, color:'var(--text-2)', lineHeight:1.5 }}>Signal-first sequences are pulling <b style={{ color:'var(--text-1)' }}>4.6× the reply rate</b> of your old templates. The biggest drop-off is reply → meeting — want me to add a ChampVoice step on hot positives (consent-gated)?</div>
        </div>
      </div>
    </div>
  )
}

Object.assign(window, { MailPage, AnalyticsPage })

/* ======== Settings.jsx ======== */
/* ChampIQ (Sim) — Settings: consent/DND capabilities, API keys, design tokens, team. */

function SettingRow({ label, desc, on, onToggle, danger }) {
  return (
    <div style={{ display:'flex', alignItems:'center', gap:14, padding:'13px 16px', borderBottom:'1px solid var(--border-1)' }}>
      <div style={{ flex:1 }}>
        <div style={{ fontFamily:'var(--font-mono)', fontSize:12.5, color:danger&&on?'var(--warn)':'var(--text-1)' }}>{label}</div>
        <div style={{ fontSize:11.5, color:'var(--text-3)', marginTop:2 }}>{desc}</div>
      </div>
      <Switch checked={on} onChange={onToggle} />
    </div>
  )
}

const ACCENT_THEMES = [
  { id:'violet', hex:'#7C5CFF' }, { id:'mint', hex:'#00E5C7' }, { id:'coral', hex:'#FF7A59' }, { id:'sun', hex:'#FFC23F' }, { id:'sky', hex:'#5BC0FF' },
]

function SettingsPage({ client, accent, setAccent }) {
  const [caps, setCaps] = React.useState(() => Object.fromEntries(CAPABILITIES.map(c=>[c.id,c.on])))
  return (
    <div className="scroll" style={{ flex:1, minWidth:0, overflowY:'auto', padding:'24px 28px' }}>
      <h2 style={{ margin:'0 0 4px', fontFamily:'var(--font-display)', fontSize:20, fontWeight:600 }}>Settings</h2>
      <div style={{ fontFamily:'var(--font-mono)', fontSize:10.5, color:'var(--text-3)', marginBottom:22, letterSpacing:'.06em' }}>{client.name.toUpperCase()} · WORKSPACE & GUARDRAILS</div>

      <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:18, alignItems:'start' }}>
        {/* Capabilities */}
        <div>
          <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}><Icon name="user" size={14} stroke="var(--accent-1)" /><span style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>Consent & DND capabilities</span></div>
          <div style={{ fontSize:12, color:'var(--text-3)', marginBottom:10 }}>The zerolang guardrail reads these. A canvas node can't ship without its capability declared here.</div>
          <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, overflow:'hidden' }}>
            {CAPABILITIES.map(c => <SettingRow key={c.id} label={c.label} desc={c.desc} danger={c.id==='auto.voice'} on={caps[c.id]} onToggle={v=>setCaps(s=>({...s,[c.id]:v}))} />)}
          </div>
        </div>

        {/* API keys + tokens + team */}
        <div style={{ display:'flex', flexDirection:'column', gap:18 }}>
          <div>
            <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}><Icon name="key" size={14} stroke="var(--accent-1)" /><span style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>API keys & providers</span></div>
            <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, overflow:'hidden' }}>
              {API_KEYS.map((k,i)=>(
                <div key={k.svc} style={{ display:'flex', alignItems:'center', gap:12, padding:'12px 16px', borderBottom:i<API_KEYS.length-1?'1px solid var(--border-1)':'none' }}>
                  <span style={{ width:7, height:7, borderRadius:'50%', background:'var(--success)', boxShadow:'0 0 6px var(--success)' }} />
                  <div style={{ width:80, flexShrink:0, fontFamily:'var(--font-display)', fontWeight:600, fontSize:12.5 }}>{k.svc}</div>
                  <div style={{ flex:1, fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-3)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{k.val}</div>
                  <button style={{ width:26, height:26, display:'grid', placeItems:'center', borderRadius:6, border:'1px solid var(--border-1)', background:'var(--bg-2)', color:'var(--text-3)', cursor:'pointer' }}><Icon name="refresh" size={13} /></button>
                </div>
              ))}
            </div>
          </div>

          <div>
            <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}><Icon name="palette" size={14} stroke="var(--accent-1)" /><span style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>Accent theme</span></div>
            <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, padding:'16px', display:'flex', gap:12, alignItems:'center' }}>
              {ACCENT_THEMES.map(th => {
                const on = accent===th.id
                return <button key={th.id} onClick={()=>setAccent(th.id)} title={th.id} style={{ width:34, height:34, borderRadius:'50%', background:th.hex, border:on?'2px solid var(--text-1)':'2px solid transparent', boxShadow:on?`0 0 14px -2px ${th.hex}`:'none', cursor:'pointer', outline:'none' }} />
              })}
              <span style={{ fontFamily:'var(--font-mono)', fontSize:11, color:'var(--text-3)', marginLeft:'auto' }}>--accent-2 · live</span>
            </div>
          </div>

          <div>
            <div style={{ display:'flex', alignItems:'center', gap:8, marginBottom:10 }}><Icon name="users" size={14} stroke="var(--accent-1)" /><span style={{ fontFamily:'var(--font-display)', fontSize:14, fontWeight:600 }}>Team & roles</span></div>
            <div style={{ background:'var(--bg-1)', border:'1px solid var(--border-1)', borderRadius:12, overflow:'hidden' }}>
              {TEAM.map((m,i)=>(
                <div key={m.name} style={{ display:'flex', alignItems:'center', gap:12, padding:'11px 16px', borderBottom:i<TEAM.length-1?'1px solid var(--border-1)':'none' }}>
                  <div style={{ width:30, height:30, borderRadius:'50%', background:`color-mix(in oklch, ${m.accent} 30%, var(--bg-2))`, border:`1.5px solid ${m.accent}`, display:'grid', placeItems:'center', color:m.accent, fontFamily:'var(--font-display)', fontWeight:700, fontSize:12 }}>{m.initial}</div>
                  <div style={{ flex:1, fontFamily:'var(--font-display)', fontWeight:600, fontSize:13 }}>{m.name}</div>
                  <Tag color={m.role==='Owner'?'var(--accent-2)':m.role==='Operator'?'var(--mint-2)':'var(--text-4)'}>{m.role}</Tag>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

window.SettingsPage = SettingsPage

/* ======== app.jsx ======== */
/* ChampIQ (Sim) — app: login gate + client state + 8-page router + Copilot host. */
/* useHashRoute, Rail, TopBar are top-level fns from shell (concatenated scope) */

const TITLES = { hub:'Dashboard', canvas:'Canvas', graph:'ChampGraph', sequences:'Sequences', personalize:'Personalize', inbox:'Inbox', prospects:'Prospects', deliver:'Deliverability', analytics:'Analytics', settings:'Settings' }

function App() {
  const [route, sub, go] = useHashRoute()
  const [authed, setAuthed] = React.useState(() => { try { return localStorage.getItem('cq:authed')==='1' } catch { return false } })
  const [clientId, setClientId] = React.useState(() => { try { return localStorage.getItem('cq:client') || 'acme' } catch { return 'acme' } })
  const [accent, setAccent] = React.useState(() => { try { return localStorage.getItem('cq:accent') || 'violet' } catch { return 'violet' } })
  const [activeCanvas, setActiveCanvas] = React.useState(null)
  const [copilot, setCopilot] = React.useState(false)
  const client = CLIENTS.find(c=>c.id===clientId) || CLIENTS[0]

  React.useEffect(()=>{ try { localStorage.setItem('cq:client', clientId) } catch {} }, [clientId])
  React.useEffect(()=>{ try { localStorage.setItem('cq:accent', accent) } catch {} }, [accent])
  React.useEffect(()=>{ try { localStorage.setItem('cq:authed', authed?'1':'0') } catch {} }, [authed])

  function enter(cid) { setClientId(cid); setAuthed(true); go('hub') }
  function openCanvas(c) { setActiveCanvas(c); go('canvas') }
  function newCanvas(t) {
    const g = CANVAS_GRAPHS[t.graph] || CANVAS_GRAPHS.sdr
    setActiveCanvas({ id:'new', name:t.title, client:clientId, nodes:g.nodes.length, graph:{ nodes:g.nodes.map(n=>({...n})), edges:g.edges.map(e=>({...e})) } })
    go('canvas')
  }

  // Login gate
  if (!authed) {
    return <div data-accent={accent} style={{ width:'100%', height:'100%' }}><LoginScreen onEnter={enter} /></div>
  }

  const canvasForRoute = activeCanvas || CANVASES.find(c=>c.client===clientId) || CANVASES[0]

  let page
  if (route === 'canvas') page = <CanvasPage canvas={canvasForRoute} onCopilot={()=>setCopilot(true)} />
  else if (route === 'graph') page = <GraphPage client={client} onCopilot={()=>setCopilot(true)} />
  else if (route === 'inbox') page = <InboxPage client={client} onCopilot={()=>setCopilot(true)} />
  else if (route === 'sequences') page = <SequencePage onPersonalize={()=>go('personalize')} />
  else if (route === 'personalize') page = <PersonalizePage client={client} onOpenInbox={()=>go('inbox')} />
  else if (route === 'prospects') page = <ProspectsPage client={client} onOpenGraph={()=>go('graph')} onEnroll={()=>go('sequences')} />
  else if (route === 'deliver') page = <DeliverabilityPage client={client} />
  else if (route === 'analytics') page = <AnalyticsPage client={client} />
  else if (route === 'settings') page = <SettingsPage client={client} accent={accent} setAccent={setAccent} />
  else page = <HubPage client={client} onOpenCanvas={openCanvas} onNewCanvas={newCanvas} onCopilot={()=>setCopilot(true)} go={go} />

  return (
    <div data-accent={accent} style={{ width:'100%', height:'100%', display:'flex', background:'var(--bg-0)', color:'var(--text-1)', fontFamily:'var(--font-body)', overflow:'hidden' }}>
      <Rail route={route} go={go} />
      <div style={{ flex:1, display:'flex', flexDirection:'column', minWidth:0, position:'relative' }}>
        <TopBar client={client} setClient={setClientId} onCopilot={()=>setCopilot(true)} title={TITLES[route] || 'Dashboard'} />
        {page}
        <Copilot open={copilot} onClose={()=>setCopilot(false)} />
      </div>
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />)
