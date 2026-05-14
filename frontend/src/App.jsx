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
          <th className="px-6 py-3 rounded-l-lg">State / Category</th>
          <th className="px-6 py-3 text-right">Total Impressions</th>
          <th className="px-6 py-3 rounded-r-lg">Relative Volume</th>
        </tr>
      </thead>
      <tbody>
        {data && data.length > 0 ? data.map((item, index) => {
          const maxVal = Math.max(...data.map(s => s.total_impressions || 0));
          const width = maxVal > 0 ? (item.total_impressions / maxVal) * 100 : 0;
          return (
            <tr key={index} className="border-b border-slate-700/50 hover:bg-slate-700/20 transition-colors">
              <td className="px-6 py-4 font-medium text-white">{item.state}</td>
              <td className="px-6 py-4 text-right font-mono">{item.total_impressions?.toLocaleString()}</td>
              <td className="px-6 py-4">
                <div className="w-full bg-slate-700 rounded-full h-1.5">
                  <div className="bg-blue-500 h-1.5 rounded-full" style={{ width: `${width}%`, transition: 'width 0.5s ease' }} />
                </div>
              </td>
            </tr>
          );
        }) : (
          <tr>
            <td colSpan="3" className="px-6 py-8 text-center text-slate-500 italic">No reports from Python (8081) yet...</td>
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
    reportingData: [], 
  });

  useEffect(() => {
    const fetchAllData = async () => {
      try {
        // 1. Fetch de Real-time (Go API - Puerto 8085)
        const statsResponse = await fetch('http://localhost:8085/api/stats');
        const statsJson = await statsResponse.json();

        // 2. Fetch de Reporting (Python FastAPI - Puerto 8081)
        let reportJson = [];
        try {
          // RUTA CORREGIDA: Apunta al endpoint de Python
          const reportResponse = await fetch('http://localhost:8081/report/top-states');
          if (reportResponse.ok) {
            reportJson = await reportResponse.json();
          }
        } catch (e) {
          console.warn("Python Reporting service (8081) is not responding.");
        }

        setData(prev => {
          const ctr = statsJson.impressions > 0 ? ((statsJson.clicks / statsJson.impressions) * 100).toFixed(1) : 0;
          const convRate = statsJson.clicks > 0 ? ((statsJson.conversions / statsJson.clicks) * 100).toFixed(1) : 0;
          
          const newTime = new Date().toLocaleTimeString().split(' ')[0];
          const newHistory = [...prev.history, { time: newTime, val: statsJson.impressions }].slice(-20);

          return {
            ...statsJson,
            ctr,
            convRate,
            history: newHistory,
            // Guardamos la lista de estados del 8081
            reportingData: reportJson.length > 0 ? reportJson : prev.reportingData
          };
        });

      } catch (error) {
        console.error("Critical error fetching data:", error);
      }
    };

    // Actualización cada 1.5 segundos para fluidez
    const interval = setInterval(fetchAllData, 1500); 
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
          <p className="text-slate-400 text-sm mt-1">Multi-source: Real-time (8085) & DuckDB Reports (8081)</p>
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
              <h2 className="text-sm font-semibold uppercase tracking-wider">Throughput (Impressions)</h2>
            </div>
            <span className="text-[10px] bg-slate-700 px-2 py-1 rounded text-slate-400 font-mono">SOURCE: 8085</span>
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
                  dot={false} 
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
            <div className="w-px h-16 bg-slate-700" />
            <GaugeChart value={data.convRate} label="CR" color="#10b981" />
          </div>
        </div>

        <div className="lg:col-span-2 bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-2 text-slate-400">
              <Map size={18} />
              <h2 className="text-sm font-semibold uppercase tracking-wider">Reports Summary (DuckDB)</h2>
            </div>
            <span className="text-[10px] bg-slate-700 px-2 py-1 rounded text-slate-400 font-mono">SOURCE: 8081</span>
          </div>
          {/* Tabla que consume directamente del servicio de Python */}
          <StateTable data={data.reportingData} />
        </div>
      </div>
    </div>
  );
};

export default App;