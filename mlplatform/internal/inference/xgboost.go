package inference

import (
	"context"

	"github.com/dmitryikh/leaves"
)

// XGBoostModel uses pure-Go tree evaluation (no CGo).
type XGBoostModel struct {
	name         string
	version      string
	ensemble     *leaves.Ensemble
	featureOrder []string
}

func LoadXGBoost(name, version, path string, featureNames []string) (*XGBoostModel, error) {
	ensemble, err := leaves.XGEnsembleFromFile(path, false)
	if err != nil {
		return nil, err
	}
	return &XGBoostModel{
		name:         name,
		version:      version,
		ensemble:     ensemble,
		featureOrder: featureNames,
	}, nil
}

func (m *XGBoostModel) Name() string    { return m.name }
func (m *XGBoostModel) Version() string { return m.version }

func (m *XGBoostModel) Predict(ctx context.Context, features map[string]float64) (float64, error) {
	row := make([]float64, len(m.featureOrder))
	for i, name := range m.featureOrder {
		row[i] = features[name]
	}
	score := m.ensemble.PredictSingle(row, 0)
	return score, nil
}

func (m *XGBoostModel) PredictBatch(ctx context.Context, batch []map[string]float64) ([]float64, error) {
	results := make([]float64, len(batch))
	for i, features := range batch {
		score, err := m.Predict(ctx, features)
		if err != nil {
			return nil, err
		}
		results[i] = score
	}
	return results, nil
}

func (m *XGBoostModel) Close() error { return nil }
