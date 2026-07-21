import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';
import { ShieldAlert, Activity, Cpu } from 'lucide-react';
import './index.css';

export default function App() {
  const [chartData, setChartData] = useState([]);
  const [internalMetrics, setInternalMetrics] = useState({ friction: 0, ckScore: 0, status: 'SAFE', alpha: 5.0, entropy: 10, actionMsg: '' });
  const [externalTargets, setExternalTargets] = useState([]);
  
  const [ws, setWs] = useState(null);
  const [scenario, setScenario] = useState('NORMAL');
  const [newUrl, setNewUrl] = useState("");
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isEngineActive, setIsEngineActive] = useState(false);
  
  const [savedTargets, setSavedTargets] = useState(() => {
    try {
      return JSON.parse(localStorage.getItem('ouroboros_targets')) || [];
    } catch {
      return [];
    }
  });

  useEffect(() => {
    localStorage.setItem('ouroboros_targets', JSON.stringify(savedTargets));
  }, [savedTargets]);

  useEffect(() => {
    const websocket = new WebSocket('ws://localhost:8082/ws');
    
    websocket.onopen = () => setWs(websocket);

    websocket.onmessage = (event) => {
      const payload = JSON.parse(event.data);
      
      if (payload.engineActive !== undefined) {
        setIsEngineActive(payload.engineActive);
      }

      // Update Internal Chart Data
      if (payload.internal && payload.internal.time) {
        setInternalMetrics(payload.internal);
        setChartData(prev => {
          const next = [...prev, payload.internal];
          if (next.length > 20) return next.slice(1);
          return next;
        });
      }

      // Update External API Data
      if (payload.external) {
        setExternalTargets(prev => {
          return payload.external.map(newTarget => {
            const existing = prev.find(t => t.url === newTarget.url);
            const prevHistory = existing && existing.history ? existing.history : [];
            return {
              ...newTarget,
              history: [...prevHistory, { friction: newTarget.friction, latency: newTarget.latency }].slice(-30)
            };
          });
        });
      }
    };

    return () => websocket.close();
  }, []);

  const handleScenarioChange = (newScenario) => {
    setScenario(newScenario);
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ scenario: newScenario }));
    }
  };

  const toggleEngine = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ action: "TOGGLE_ENGINE", enabled: !isEngineActive }));
    }
  };

  const addTarget = (urlToStart) => {
    if (ws && ws.readyState === WebSocket.OPEN && urlToStart) {
      ws.send(JSON.stringify({ action: "ADD_TARGET", url: urlToStart }));
    }
  };

  const saveAndStartTarget = () => {
    if (newUrl && !savedTargets.includes(newUrl)) {
      setSavedTargets([...savedTargets, newUrl]);
    }
    addTarget(newUrl);
    setNewUrl("");
  };

  const removeTarget = (urlToRemove) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ action: "REMOVE_TARGET", url: urlToRemove }));
    }
  };

  return (
    <div className="dashboard-container" style={{maxWidth: '1200px', margin: '0 auto', paddingBottom: '4rem'}}>
      
      {/* HEADER GERAL */}
      <header className="header" style={{ borderBottom: '1px solid rgba(255,255,255,0.05)', paddingBottom: '1rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div className="title-section">
          <h1>Genesis Reflexivity</h1>
          <p>Ouroboros Hybrid Architecture (Internal + External Sentinel)</p>
          <div style={{marginTop: '1rem', display: 'flex', gap: '10px'}}>
             <button onClick={() => setActiveTab('dashboard')} style={{background: activeTab === 'dashboard' ? '#b485ff' : '#2a2e38', color: '#fff', border: 'none', padding: '6px 15px', borderRadius: '4px', cursor: 'pointer'}}>Painel Principal</button>
             <button onClick={() => setActiveTab('history')} style={{background: activeTab === 'history' ? '#b485ff' : '#2a2e38', color: '#fff', border: 'none', padding: '6px 15px', borderRadius: '4px', cursor: 'pointer'}}>Tabela de Auditoria (DB)</button>
             <button onClick={toggleEngine} style={{
                background: isEngineActive ? 'rgba(0, 240, 255, 0.2)' : 'rgba(255, 51, 102, 0.2)',
                color: isEngineActive ? '#00f0ff' : '#ff3366',
                border: `1px solid ${isEngineActive ? '#00f0ff' : '#ff3366'}`,
                padding: '6px 15px', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold'
              }}>
                Motor Relacional: {isEngineActive ? 'ON' : 'OFF'}
             </button>
          </div>
        </div>
        <div>
          {ws && ws.readyState === WebSocket.OPEN ? (
            <div className="status-badge status-SAFE">SYSTEM ONLINE</div>
          ) : (
            <div className="status-badge status-DENY">SYSTEM OFFLINE</div>
          )}
        </div>
      </header>

      {activeTab === 'dashboard' ? (
        <>
          {/* SESSÃO 1: SIMULADOR INTERNO E GRÁFICO */}
          <h2 style={{ color: '#fff', marginTop: '2rem' }}>1. Monitor Interno (GenesisCore Kernel)</h2>
      
      {internalMetrics.actionMsg && (
        <div style={{
          backgroundColor: 'rgba(255, 51, 102, 0.2)',
          border: '1px solid #ff3366', color: '#ff3366', padding: '1rem',
          borderRadius: '8px', marginBottom: '1.5rem', fontWeight: 'bold', textAlign: 'center',
          animation: 'pulse 1s infinite'
        }}>
          {internalMetrics.actionMsg}
        </div>
      )}

      <div className="grid-3" style={{ marginBottom: '1rem' }}>
        <div className="glass-card">
          <div className="metric-label"><Cpu size={16}/> Network Entropy (Simulated)</div>
          <p className="metric-value">{internalMetrics.entropy} <span className="metric-unit">ops/s</span></p>
        </div>
        <div className="glass-card">
          <div className="metric-label"><Activity size={16}/> Expected Alpha (α)</div>
          <p className="metric-value">{internalMetrics.alpha} <span className="metric-unit">baseline</span></p>
        </div>
        <div className="glass-card" style={{
            borderColor: internalMetrics.status === 'DENY' ? 'rgba(255, 51, 102, 0.5)' : 
                         internalMetrics.status === 'WARN' ? 'rgba(255, 171, 0, 0.5)' : 'rgba(255,255,255,0.05)'
        }}>
          <div className="metric-label"><ShieldAlert size={16}/> Final Friction</div>
          <p className="metric-value">{internalMetrics.friction} <span className="metric-unit">/ 100</span></p>
        </div>
      </div>

      <div className="glass-card" style={{ height: '350px', paddingBottom: '3rem', marginBottom: '1rem' }}>
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
            <XAxis dataKey="time" stroke="#8892b0" fontSize={12} tickMargin={10} />
            <YAxis domain={[0, 100]} stroke="#8892b0" fontSize={12} />
            <Tooltip contentStyle={{backgroundColor: '#0d0f12', borderColor: 'rgba(255,255,255,0.1)'}} />
            <ReferenceLine y={20} stroke="rgba(255, 171, 0, 0.3)" strokeDasharray="3 3" />
            <ReferenceLine y={70} stroke="rgba(255, 51, 102, 0.3)" strokeDasharray="3 3" />
            <Line type="monotone" dataKey="friction" stroke="#00f0ff" strokeWidth={3} dot={false} isAnimationActive={false} name="Fricção Atual" />
            <Line type="step" dataKey="alpha" stroke="#b485ff" strokeWidth={2} strokeDasharray="5 5" dot={false} isAnimationActive={false} name="Vetor Alpha" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="glass-card" style={{ marginBottom: '3rem' }}>
        <h3 style={{marginTop: 0, marginBottom: '1rem', fontSize: '1rem'}}>Painel de Teste de Estresse</h3>
        <div className="controls">
          <button onClick={() => handleScenarioChange('NORMAL')} style={{borderColor: scenario === 'NORMAL' ? '#00ff88' : ''}}>Operação Normal</button>
          <button onClick={() => handleScenarioChange('HIGH_LOAD_EXPECTED')} style={{borderColor: scenario === 'HIGH_LOAD_EXPECTED' ? '#ffab00' : ''}}>Carga Alta</button>
          <button className="btn-danger" onClick={() => handleScenarioChange('ATTACK')} style={{borderColor: scenario === 'ATTACK' ? '#ff3366' : ''}}>Injetar Ataque</button>
        </div>
      </div>

      {/* SESSÃO 2: SENTINELA DINÂMICO (APIs EXTERNAS) */}
      <h2 style={{ color: '#fff', borderTop: '1px solid rgba(255,255,255,0.1)', paddingTop: '2rem' }}>2. Sentinela Extracorpóreo (API Probing Real)</h2>
      
      <div style={{ display: 'flex', gap: '10px', marginBottom: '1rem' }}>
        <input 
          value={newUrl} 
          onChange={e => setNewUrl(e.target.value)} 
          placeholder="https://api.exemplo.com"
          style={{ padding: '10px 16px', borderRadius: '6px', border: '1px solid rgba(255,255,255,0.2)', background: '#12141a', color: '#fff', width: '400px' }}
        />
        <button onClick={saveAndStartTarget} style={{ padding: '10px 20px', background: '#b485ff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold' }}>
          Salvar & Monitorar
        </button>
      </div>

      <div style={{marginBottom: '2rem'}}>
        <h3 style={{color: '#8892b0', fontSize: '0.9rem', textTransform: 'uppercase'}}>Trilhas Salvas (Cache)</h3>
        {savedTargets.length === 0 ? <p style={{color: '#555'}}>Nenhuma API salva no cache.</p> : null}
        <div style={{display: 'flex', flexWrap: 'wrap', gap: '10px'}}>
          {savedTargets.map(url => {
            const isRunning = externalTargets.some(t => t.url === url);
            return (
              <div key={url} style={{
                background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)',
                padding: '10px 15px', borderRadius: '6px', display: 'flex', alignItems: 'center', gap: '15px'
              }}>
                <span style={{fontFamily: 'monospace', color: '#fff', fontSize: '0.9rem'}}>{url}</span>
                {isRunning ? (
                  <button onClick={() => removeTarget(url)} style={{background: 'rgba(255, 51, 102, 0.2)', color: '#ff3366', border: 'none', borderRadius: '4px', padding: '4px 8px', cursor: 'pointer', fontSize: '0.8rem'}}>⏸ Parar</button>
                ) : (
                  <button onClick={() => addTarget(url)} style={{background: 'rgba(0, 240, 255, 0.2)', color: '#00f0ff', border: 'none', borderRadius: '4px', padding: '4px 8px', cursor: 'pointer', fontSize: '0.8rem'}}>▶ Iniciar</button>
                )}
                <button onClick={() => {
                   setSavedTargets(prev => prev.filter(u => u !== url));
                   if (isRunning) removeTarget(url);
                }} style={{background: 'transparent', color: '#8892b0', border: 'none', cursor: 'pointer', fontSize: '1rem'}}>🗑</button>
              </div>
            )
          })}
        </div>
      </div>

      {externalTargets.length === 0 ? (
        <p style={{ color: '#8892b0' }}>Nenhuma API externa sendo monitorada ainda.</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {externalTargets.map(t => (
            <div key={t.url} className="glass-card" style={{ 
              position: 'relative',
              borderColor: t.status === 'DENY' ? 'rgba(255, 51, 102, 0.5)' : t.status === 'WARN' ? 'rgba(255, 171, 0, 0.5)' : 'rgba(255,255,255,0.05)',
              display: 'grid', gridTemplateColumns: '2fr 1fr 1fr 1fr 1fr', gap: '1rem', alignItems: 'center'
            }}>
              <div 
                onClick={() => removeTarget(t.url)}
                style={{ position: 'absolute', top: '10px', right: '10px', cursor: 'pointer', color: '#ff3366', fontSize: '0.8rem', padding: '4px 8px', background: 'rgba(255, 51, 102, 0.1)', borderRadius: '4px' }}
              >
                ⏸ Parar
              </div>
              <div>
                <div style={{fontSize: '0.75rem', color: '#8892b0', textTransform: 'uppercase'}}>Alvo (API)</div>
                <div style={{fontFamily: 'monospace', color: '#fff', wordBreak: 'break-all'}}>{t.url}</div>
                {t.actionMsg && <div style={{color: '#ff3366', fontSize: '0.8rem', marginTop: '5px', animation: 'pulse 1s infinite'}}>{t.actionMsg}</div>}
              </div>
              <div>
                <div style={{fontSize: '0.75rem', color: '#8892b0'}}>LATÊNCIA REAL</div>
                <div style={{fontSize: '1.4rem', color: '#fff'}}>{t.latency} <span style={{fontSize: '0.8rem'}}>ms</span></div>
              </div>
              <div>
                <div style={{fontSize: '0.75rem', color: '#8892b0'}}>ALPHA</div>
                <div style={{fontSize: '1.4rem', color: '#b485ff'}}>{t.alpha}</div>
              </div>
              <div>
                <div style={{fontSize: '0.75rem', color: '#8892b0'}}>FRICÇÃO</div>
                <div style={{fontSize: '1.4rem', color: t.status === 'DENY' ? '#ff3366' : '#fff'}}>{t.friction}</div>
              </div>
              <div className={`status-badge status-${t.status}`} style={{justifySelf: 'end'}}>
                {t.status}
              </div>
              
              <div style={{ gridColumn: '1 / -1', height: '40px', marginTop: '5px' }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={t.history}>
                    <YAxis domain={['dataMin', 'dataMax']} hide />
                    <Line 
                      type="monotone" 
                      dataKey="friction" 
                      stroke={t.status === 'SAFE' ? '#00f0ff' : t.status === 'WARN' ? '#ffab00' : '#ff3366'} 
                      strokeWidth={2} 
                      dot={false} 
                      isAnimationActive={false} 
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          ))}
        </div>
      )}
        </>
      ) : (
        <HistoryView />
      )}

    </div>
  );
}

function HistoryView() {
  const [data, setData] = useState([]);
  
  useEffect(() => {
    fetch('http://localhost:8082/api/history')
      .then(r => r.json())
      .then(d => setData(d))
      .catch(e => console.error("Erro ao ler db:", e));
  }, []);
  
  return (
    <div style={{color: '#fff', marginTop: '2rem'}}>
      <h2 style={{marginBottom: '1rem'}}>Audit Log (SQLite)</h2>
      <p style={{color: '#8892b0', marginBottom: '2rem'}}>
        Exibindo as últimas 500 pontuações temporais salvas no banco de dados.
      </p>
      
      <div className="glass-card" style={{overflowX: 'auto', padding: '1rem'}}>
        <table style={{width: '100%', textAlign: 'left', borderCollapse: 'collapse'}}>
          <thead>
            <tr style={{borderBottom: '1px solid rgba(255,255,255,0.1)'}}>
              <th style={{padding: '12px'}}>ID</th>
              <th style={{padding: '12px'}}>Target URL</th>
              <th style={{padding: '12px'}}>Ck Score (Entropia)</th>
              <th style={{padding: '12px'}}>Timestamp</th>
            </tr>
          </thead>
          <tbody>
            {data.map(r => (
              <tr key={r.id} style={{borderBottom: '1px solid rgba(255,255,255,0.05)'}}>
                <td style={{padding: '12px', color: '#8892b0'}}>#{r.id}</td>
                <td style={{padding: '12px', fontFamily: 'monospace'}}>{r.serviceId}</td>
                <td style={{padding: '12px', color: '#00f0ff'}}>{r.ckScore.toFixed(4)}</td>
                <td style={{padding: '12px', color: '#b485ff'}}>{r.timestamp}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
