package inference

import (
	"context"
	"fmt"
	"sort"

	ort "github.com/yalue/onnxruntime_go"
)

type ONNXModel struct {
	name         string
	version      string
	session      *ort.DynamicAdvancedSession
	inputNames   []string
	outputNames  []string
	featureOrder []string // deterministic feature ordering
}

func LoadONNX(name, version, path string, featureNames []string) (*ONNXModel, error) {
	// Initialize ONNX Runtime shared library (call once globally)
	// ort.SetSharedLibraryPath("libonnxruntime.so")

	outputNames := []string{"prediction"}

	session, err := ort.NewDynamicAdvancedSession(path,
		[]string{"features"}, outputNames,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("load onnx %s: %w", path, err)
	}

	sorted := make([]string, len(featureNames))
	copy(sorted, featureNames)
	sort.Strings(sorted)

	return &ONNXModel{
		name:         name,
		version:      version,
		session:      session,
		inputNames:   []string{"features"},
		outputNames:  outputNames,
		featureOrder: sorted,
	}, nil
}

func (m *ONNXModel) Name() string    { return m.name }
func (m *ONNXModel) Version() string { return m.version }

func (m *ONNXModel) Predict(ctx context.Context, features map[string]float64) (float64, error) {
	results, err := m.PredictBatch(ctx, []map[string]float64{features})
	if err != nil {
		return 0, err
	}
	return results[0], nil
}

func (m *ONNXModel) PredictBatch(ctx context.Context, batch []map[string]float64) ([]float64, error) {
	batchSize := len(batch)
	numFeatures := len(m.featureOrder)

	// Flatten into a single input tensor [batchSize, numFeatures]
	inputData := make([]float32, batchSize*numFeatures)
	for i, features := range batch {
		for j, fname := range m.featureOrder {
			inputData[i*numFeatures+j] = float32(features[fname])
		}
	}

	inputShape := []int64{int64(batchSize), int64(numFeatures)}
	outputShape := []int64{int64(batchSize), 1}

	inputTensor, err := ort.NewTensor[float32](inputShape, inputData)
	if err != nil {
		return nil, err
	}
	defer inputTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, err
	}
	defer outputTensor.Destroy()

	err = m.session.Run(
		[]ort.ArbitraryTensor{inputTensor},
		[]ort.ArbitraryTensor{outputTensor},
	)
	if err != nil {
		return nil, fmt.Errorf("onnx run: %w", err)
	}

	outputData := outputTensor.GetData()
	results := make([]float64, batchSize)
	for i := range results {
		results[i] = float64(outputData[i])
	}
	return results, nil
}

func (m *ONNXModel) Close() error {
	return m.session.Destroy()
}
