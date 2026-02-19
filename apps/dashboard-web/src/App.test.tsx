import React from 'react';
import { render } from '@testing-library/react';
import App from './App';

jest.mock('react-leaflet', () => {
  const React = require('react');
  const createComponent = (name: string) => ({ children }: { children?: React.ReactNode }) =>
    React.createElement('div', { 'data-testid': `leaflet-${name}` }, children);
  return {
    MapContainer: createComponent('map-container'),
    Marker: createComponent('marker'),
    Popup: createComponent('popup'),
    TileLayer: createComponent('tile-layer'),
    Polyline: createComponent('polyline'),
    useMap: () => ({}),
    useMapEvents: () => ({}),
    useMapEvent: () => ({}),
  };
});

test('renders without crashing', () => {
  const { container } = render(<App />);
  expect(container.firstChild).toBeTruthy();
});
