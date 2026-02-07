import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../store';
import { acknowledgeAlert, fetchAlerts } from '../store/alertsSlice';
import { Alert } from '../types';

const Alerts: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { alerts, loading } = useSelector((state: RootState) => state.alerts);

  useEffect(() => {
    dispatch(fetchAlerts({}) as any);
  }, [dispatch]);

  const handleAcknowledge = (id: string) => {
    dispatch(acknowledgeAlert(id));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-lg text-gray-600">Loading alerts...</div>
      </div>
    );
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">Alerts</h1>

      <div className="space-y-4">
        {alerts.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">
            No alerts found
          </div>
        ) : (
          alerts.map((alert: Alert) => (
            <div
              key={alert.id}
              className={`bg-white rounded-lg shadow border-l-4 p-4 ${getSeverityColor(
                alert.severity
              )}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-2">
                    <span className="font-semibold text-lg">{alert.type}</span>
                    <span
                      className={`px-2 py-1 text-xs font-semibold rounded ${
                        alert.severity === 'critical'
                          ? 'bg-red-200 text-red-900'
                          : alert.severity === 'warning'
                          ? 'bg-yellow-200 text-yellow-900'
                          : 'bg-blue-200 text-blue-900'
                      }`}
                    >
                      {alert.severity.toUpperCase()}
                    </span>
                  </div>
                  <p className="text-gray-700 mb-2">{alert.message}</p>
                  <p className="text-sm text-gray-500">
                    {new Date(alert.timestamp).toLocaleString()}
                  </p>
                </div>
                {!alert.acknowledged && (
                  <button
                    onClick={() => handleAcknowledge(alert.id)}
                    className="ml-4 px-4 py-2 bg-gray-800 text-white text-sm rounded hover:bg-gray-900"
                  >
                    Acknowledge
                  </button>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Alerts;
