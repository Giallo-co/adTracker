import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Activity, MousePointer2, Target, BarChart3, Map } from 'lucide-react';

// --- COMPONENTES AUXILIARES (UI) ---

const Card = ({ title, value, icon: Icon, color }) => (
  <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
    <div className="flex justify-between items-center mb-4">
      <h3 className="text-slate-400 text-sm font-medium uppercase tracking-wider">{title}</h3>
      <Icon size={20} className={color} />
    </div>
    <p className="text-4xl font-bold text-white">{value ? value.toLocaleString() : 0}</p>
  </div>
);

const GaugeChart = ({ value, label, color }) => {
  const angle = (value / 100) * 180;
  return (
    <div className="flex flex-col items-center gap-2">
      <div className="relative w-32 h-16 overflow-hidden">
        <svg viewBox="0 0 100 50" className="w-full h-full">
          <path d="M10 50 A40 40 0 0 1 90 50" fill="none" stroke="#334155" strokeWidth="10" />
          <path 
            d="M10 50 A40 40 0 0 1 90 50" 
            fill="none" 
            stroke={color} 
            strokeWidth="10" 
            strokeDasharray={`${angle} 180`} 
            style={{ transition: 'stroke-dasharray 0.5s ease' }}
          />
        </svg>
        <div 
          className="absolute bottom-0 left-1/2 w-1 h-12 bg-white rounded-full origin-bottom" 
          style={{ 
            transform: `translateX(-50%) rotate(${angle - 90}deg)`, 
            transition: 'transform 0.5s ease' 
          }}
        />
      </div>
      <p className="text-2xl font-bold" style={{ color }}>{value}%</p>
      <p className="text-[10px] text-slate-400 uppercase font-semibold text-center">{label}</p>
    </div>
  );
};

const StateTable = ({ data }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-sm text-left text-slate-300">
      <thead className="text-xs text-slate-500 uppercase bg-slate-900/50">
        <tr>
          <th className="px-6 py-3 rounded-l-lg">State</th>
          <th className="px-6 py-3 text-right">Events</th>
          <th className="px-6 py-3 rounded-r-lg">Relative Volume</th>
        </tr>
      </thead>
      <tbody>
        {data && data.length > 0 ? data.map((state, index) => {
          const maxVal = Math.max(...data.map(s => s.value || 0));
          const width = maxVal > 0 ? (state.value / maxVal) * 100 : 0;
          return (
            <tr key={index} className="border-b border-slate-700/50 hover:bg-slate-700/20 transition-colors">
              <td className="px-6 py-4 font-medium text-white">{state.name}</td>
              <td className="px-6 py-4 text-right font-mono">{state.value}</td>
              <td className="px-6 py-4">
                <div className="w-full bg-slate-700 rounded-full h-1.5">
                  <div className="bg-emerald-500 h-1.5 rounded-full" style={{ width: `${width}%`, transition: 'width 0.5s ease' }} />
                </div>
              </td>
            </tr>
          );
        }) : (
          <tr>
            <td colSpan="3" className="px-6 py-8 text-center text-slate-500 italic">No data from backend yet...</td>
          </tr>
        )}
      </tbody>
    </table>
  </div>
);

// --- COMPONENTE PRINCIPAL ---

const App = () => {
  const [data, setData] = useState({
    impressions: 0,
    clicks: 0,
    conversions: 0,
    ctr: 0,
    convRate: 0,
    history: [],
    stateReport: [],
  });

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch('http://localhost:8085/api/stats');
        if (!response.ok) throw new Error('Network response was not ok');
        
        const newData = await response.json();
        
        setData(prev => {
          // Calculamos tasas
          const ctr = newData.impressions > 0 ? ((newData.clicks / newData.impressions) * 100).toFixed(1) : 0;
          const convRate = newData.clicks > 0 ? ((newData.conversions / newData.clicks) * 100).toFixed(1) : 0;
          
          // Generamos el punto de la gráfica (volumen actual)
          const newTime = new Date().toLocaleTimeString().split(' ')[0];
          const newHistory = [...prev.history, { time: newTime, val: newData.impressions }].slice(-20);

          return {
            ...newData,
            ctr,
            convRate,
            history: newHistory,
            // Si el backend no manda estados aún, mantenemos lo que había
            stateReport: newData.stateReport || prev.stateReport
          };
        });
      } catch (error) {
        console.error("Error fetching data from Go backend:", error);
      }
    };

    const interval = setInterval(fetchStats, 1000); // Actualiza cada 1 segundo para más fluidez
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 p-8 font-sans selection:bg-blue-500/30">
      {/* HEADER */}
      <header className="mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent tracking-tight">
            AdTracker Signal Board
          </h1>
          <p className="text-slate-400 text-sm mt-1">Real-time processing feed from Go API</p>
        </div>
        <div className="flex items-center gap-3 bg-slate-800 px-4 py-2 rounded-full border border-slate-700 shadow-inner">
          <div className="relative flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
          </div>
          <span className="text-slate-300 text-xs font-bold uppercase tracking-widest">Live Engine</span>
        </div>
      </header>

      {/* TOP ROW: KPIs + CHART */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <div className="flex flex-col gap-6">
          <Card title="Total Impressions" value={data.impressions} icon={Activity} color="text-blue-400" />
          <Card title="Total Clicks" value={data.clicks} icon={MousePointer2} color="text-yellow-400" />
          <Card title="Total Conversions" value={data.conversions} icon={Target} color="text-emerald-400" />
        </div>

        <div className="lg:col-span-2 bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg flex flex-col">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2 text-slate-400">
              <BarChart3 size={18} />
              <h2 className="text-sm font-semibold uppercase tracking-wider">Throughput (Impressions over time)</h2>
            </div>
            <span className="text-[10px] bg-slate-700 px-2 py-1 rounded text-slate-400 font-mono">PORT: 8085</span>
          </div>
          <div className="flex-grow min-h-[320px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={data.history}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} axisLine={false} />
                <YAxis stroke="#64748b" fontSize={11} tickLine={false} axisLine={false} />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#0f172a', border: '1px solid #334155', borderRadius: '8px', fontSize: '12px' }}
                  itemStyle={{ color: '#60a5fa' }}
                  cursor={{ stroke: '#334155', strokeWidth: 2 }}
                />
                <Line 
                  type="monotone" 
                  dataKey="val" 
                  stroke="#3b82f6" 
                  strokeWidth={3} 
                  dot={{ r: 0 }} 
                  activeDot={{ r: 6, strokeWidth: 0 }} 
                  isAnimationActive={false} 
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* BOTTOM ROW: GAUGES + TABLE */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-slate-800 p-8 rounded-xl border border-slate-700 shadow-lg flex flex-col justify-center items-center">
          <h2 className="text-slate-500 text-[10px] font-black uppercase self-start mb-8 tracking-[0.2em]">Efficiency Metrics</h2>
          <div className="flex justify-around w-full items-center">
            <GaugeChart value={data.ctr} label="CTR" color="#fbbf24" />
            <div className="w-px h-16 bg-slate-700 shadow-glow" />
            <GaugeChart value={data.convRate} label="CR" color="#10b981" />
          </div>
        </div>

        <div className="lg:col-span-2 bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
          <div className="flex items-center gap-2 mb-6 text-slate-400">
            <Map size={18} />
            <h2 className="text-sm font-semibold uppercase tracking-wider">Geographic Distribution</h2>
          </div>
          <StateTable data={data.stateReport} />
        </div>
      </div>
    </div>
  );
};

export default App;