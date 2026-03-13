import React, { useEffect, useState } from "react";
import {
    CartesianGrid,
    Legend,
    Line,
    LineChart,
    PolarAngleAxis,
    PolarGrid,
    PolarRadiusAxis,
    Radar,
    RadarChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import dashboardApi from "../../../services/dashboardApi";

const PolarAngleAxisAny = PolarAngleAxis as any;
const PolarRadiusAxisAny = PolarRadiusAxis as any;
const RadarChartAny = RadarChart as any;
const RadarAny = Radar as any;
const ResponsiveContainerAny = ResponsiveContainer as any;

const Analytics: React.FC = () => {
  const [efficiencyData, setEfficiencyData] = useState([]);
  const [trendData, setTrendData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        // In a real app, we'd dispatch to analyticsSlice
        const trips = (await dashboardApi.getAnalyticsTrips()) as any;
        setTrendData(trips || []);

        // Mock efficiency data
        setEfficiencyData([
          { subject: "Fuel", A: 120, B: 110, fullMark: 150 },
          { subject: "Route", A: 98, B: 130, fullMark: 150 },
          { subject: "Speed", A: 86, B: 130, fullMark: 150 },
          { subject: "Safety", A: 99, B: 100, fullMark: 150 },
          { subject: "SLA", A: 85, B: 90, fullMark: 150 },
        ] as any);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchAnalytics();
  }, []);

  return (
    <div className="p-6 space-y-6 bg-black min-h-screen text-white">
      <h1 className="text-3xl font-bold">Fleet Analytics</h1>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Fleet Efficiency Radar */}
        <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl">
          <h2 className="text-xl font-semibold mb-6">Fleet Efficiency Radar</h2>
          <div className="h-[400px]">
            <ResponsiveContainerAny width="100%" height="100%">
              <RadarChartAny
                cx="50%"
                cy="50%"
                outerRadius="80%"
                data={efficiencyData}
              >
                <PolarGrid stroke="#374151" />
                <PolarAngleAxisAny
                  dataKey="subject"
                  tick={{ fill: "#9CA3AF", fontSize: 12 }}
                />
                <PolarRadiusAxisAny stroke="#374151" />
                <RadarAny
                  name="Current Period"
                  dataKey="A"
                  stroke="#3B82F6"
                  fill="#3B82F6"
                  fillOpacity={0.5}
                />
                <RadarAny
                  name="Previous Period"
                  dataKey="B"
                  stroke="#10B981"
                  fill="#10B981"
                  fillOpacity={0.5}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#111827",
                    borderColor: "#374151",
                  }}
                />
                <Legend />
              </RadarChartAny>
            </ResponsiveContainerAny>
          </div>
        </div>

        {/* Trip Trends */}
        <div className="bg-gray-900 border border-gray-800 p-6 rounded-xl">
          <h2 className="text-xl font-semibold mb-6">Execution Trends</h2>
          <div className="h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trendData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis dataKey="name" stroke="#9CA3AF" />
                <YAxis stroke="#9CA3AF" />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "#111827",
                    borderColor: "#374151",
                  }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="completed"
                  stroke="#10B981"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                />
                <Line
                  type="monotone"
                  dataKey="delayed"
                  stroke="#EF4444"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Analytics;
