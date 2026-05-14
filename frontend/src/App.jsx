import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Activity, MousePointer2, Target, BarChart3, Map } from 'lucide-react';

// --- COMPONENTES AUXILIARES ---

const Card = ({ title, value, icon: Icon, color }) => (
  <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
    <div className="flex justify-between items-center mb-4">
      <h3 className="text-slate-400 text-sm font-medium uppercase tracking-wider">{title}</h3>
      <Icon size={20} className={color} />
    </div>
    <p className="text-4xl font-bold text-white">{value.toLocaleString()}</p>
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
        {data.map((state, index) => {
          const maxVal = Math.max(...data.map(s => s.value));
          const width = maxVal > 0 ? (state.value / maxVal) * 100 : 0;
          return (
            <tr key={index} className="border-b border-slate-700/50 hover:bg-slate-700/20 transition-colors">
              <td className="px-6 py-4 font-medium text-white">{state.name}</td>
              <td className="px-6 py-4 text-right font-mono">{state.value}</td>
              <td className="px-6 py-4">
                <div className="w-full bg-slate-700 rounded-full h-1.5">
                  <div className="bg-emerald-500 h-1.5 rounded-full" style={{ width: `${width}%` }} />
                </div>
              </td>
            </tr>
          );
        })}
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
    stateReport: [
      { name: 'California (CA)', value: 0 },
      { name: 'New York (NY)', value: 0 },
      { name: 'Texas (TX)', value: 0 },
      { name: 'Florida (FL)', value: 0 },
      { name: 'Illinois (IL)', value: 0 },
    ],
  });

  useEffect(() => {
    const interval = setInterval(() => {
      setData(prev => {
        const newImp = prev.impressions + Math.floor(Math.random() * 10) + 5;
        const newClicks = prev.clicks + (Math.random() > 0.6 ? Math.floor(Math.random() * 3) : 0);
        const newConv = prev.conversions + (Math.random() > 0.85 ? 1 : 0);
        const newTime = new Date().toLocaleTimeString().split(' ')[0];

        return {
          impressions: newImp,
          clicks: newClicks,
          conversions: newConv,
          ctr: newImp > 0 ? ((newClicks / newImp) * 100).toFixed(1) : 0,
          convRate: newClicks > 0 ? ((newConv / newClicks) * 100).toFixed(1) : 0,
          history: [...prev.history, { time: newTime, val: newImp }].slice(-15),
          stateReport: prev.stateReport.map(s => ({ ...s, value: s.value + Math.floor(Math.random() * 3) }))
        };
      });
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 p-8">
      {/* HEADER */}
      <header className="mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">
            AdTracker Signal Board
          </h1>
          <p className="text-slate-400">Personal Dashboard & Analytics System</p>
        </div>
        <div className="flex items-center gap-2 bg-emerald-500/10 px-4 py-2 rounded-full border border-emerald-500/20 text-emerald-500 text-sm font-medium animate-pulse">
          <div className="w-2 h-2 bg-emerald-500 rounded-full" />
          Live Connection Active
        </div>
      </header>

      {/* TOP ROW: KPIs + CHART */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <div className="flex flex-col gap-6">
          <Card title="Cumulative Impressions" value={data.impressions} icon={Activity} color="text-blue-400" />
          <Card title="Cumulative Clicks" value={data.clicks} icon={MousePointer2} color="text-yellow-400" />
          <Card title="Cumulative Conversions" value={data.conversions} icon={Target} color="text-emerald-400" />
        </div>

        <div className="lg:col-span-2 bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg flex flex-col">
          <div className="flex items-center gap-2 mb-6 text-slate-400 italic">
            <BarChart3 size={18} />
            <h2 className="text-sm font-semibold uppercase">Event Volume (Total Load)</h2>
          </div>
          <div className="flex-grow min-h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={data.history}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                <XAxis dataKey="time" stroke="#94a3b8" fontSize={12} tickLine={false} />
                <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
                  itemStyle={{ color: '#60a5fa' }}
                />
                <Line type="monotone" dataKey="val" stroke="#3b82f6" strokeWidth={3} dot={false} isAnimationActive={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* BOTTOM ROW: GAUGES + TABLE */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg flex flex-col justify-center items-center gap-8">
          <h2 className="text-slate-400 text-xs font-bold uppercase self-start mb-2 italic">Performance Rates</h2>
          <div className="flex justify-around w-full gap-4">
            <GaugeChart value={data.ctr} label="CTR" color="#facc15" />
            <GaugeChart value={data.convRate} label="Conv. Rate" color="#10b981" />
          </div>
        </div>

        <div className="lg:col-span-2 bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
          <div className="flex items-center gap-2 mb-6 text-slate-400 italic">
            <Map size={18} />
            <h2 className="text-sm font-semibold uppercase">Distribution by State (Live Map Data)</h2>
          </div>
          <StateTable data={data.stateReport} />
        </div>
      </div>
    </div>
  );
};

export default App;