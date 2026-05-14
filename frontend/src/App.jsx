import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Activity, MousePointer2, Target, BarChart3 } from 'lucide-react';

const App = () => {
  const [data, setData] = useState({
    impressions: 0,
    clicks: 0,
    conversions: 0,
    history: []
  });

  // Simulador temporal de datos en tiempo real (lo conectaremos a Go en el siguiente paso)
  useEffect(() => {
    const interval = setInterval(() => {
      setData(prev => {
        const newImp = prev.impressions + Math.floor(Math.random() * 5);
        const newClicks = prev.clicks + (Math.random() > 0.7 ? 1 : 0);
        const newTime = new Date().toLocaleTimeString().split(' ')[0];
        
        return {
          impressions: newImp,
          clicks: newClicks,
          conversions: prev.conversions + (Math.random() > 0.9 ? 1 : 0),
          history: [...prev.history, { time: newTime, val: newImp }].slice(-15)
        };
      });
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  const Card = ({ title, value, icon: Icon, color }) => (
    <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-slate-400 text-sm font-medium uppercase tracking-wider">{title}</h3>
        <Icon size={20} className={color} />
      </div>
      <p className="text-4xl font-bold text-white">{value.toLocaleString()}</p>
    </div>
  );

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 p-8">
      <header className="mb-10 flex justify-between items-center border-b border-slate-800 pb-6">
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">
            AdTracker Signal Board
          </h1>
          <p className="text-slate-400">Monitoreo de eventos en tiempo real</p>
        </div>
        <div className="flex items-center gap-2 bg-emerald-500/10 px-4 py-2 rounded-full border border-emerald-500/20">
          <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
          <span className="text-emerald-500 text-sm font-medium">Live Feed</span>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <Card title="Cumulative Impressions" value={data.impressions} icon={Activity} color="text-blue-400" />
        <Card title="Cumulative Clicks" value={data.clicks} icon={MousePointer2} color="text-yellow-400" />
        <Card title="Cumulative Conversions" value={data.conversions} icon={Target} color="text-emerald-400" />
      </div>

      <div className="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-lg">
        <div className="flex items-center gap-2 mb-6 text-slate-400 italic">
          <BarChart3 size={18} />
          <h2 className="text-sm font-semibold uppercase">Event Volume per Minute</h2>
        </div>
        <div className="h-80 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data.history}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
              <XAxis dataKey="time" stroke="#94a3b8" fontSize={12} tickLine={false} />
              <YAxis stroke="#94a3b8" fontSize={12} tickLine={false} />
              <Tooltip 
                contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
                itemStyle={{ color: '#60a5fa' }}
              />
              <Line 
                type="monotone" 
                dataKey="val" 
                stroke="#3b82f6" 
                strokeWidth={3} 
                dot={{ r: 4, fill: '#3b82f6' }} 
                activeDot={{ r: 6 }} 
                isAnimationActive={false} 
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};

export default App;