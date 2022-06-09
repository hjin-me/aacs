// Code generated by MockGen. DO NOT EDIT.
// Source: ../../../../vendor/go.opentelemetry.io/otel/sdk/metric/export/metric.go

// Package myoteltest is a generated GoMock package.
package myoteltest

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	instrumentation "go.opentelemetry.io/otel/sdk/instrumentation"
	aggregator "go.opentelemetry.io/otel/sdk/metric/aggregator"
	export "go.opentelemetry.io/otel/sdk/metric/export"
	aggregation "go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	sdkapi "go.opentelemetry.io/otel/sdk/metric/sdkapi"
	resource "go.opentelemetry.io/otel/sdk/resource"
)

// MockProcessor is a mock of Processor interface.
type MockProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockProcessorMockRecorder
}

// MockProcessorMockRecorder is the mock recorder for MockProcessor.
type MockProcessorMockRecorder struct {
	mock *MockProcessor
}

// NewMockProcessor creates a new mock instance.
func NewMockProcessor(ctrl *gomock.Controller) *MockProcessor {
	mock := &MockProcessor{ctrl: ctrl}
	mock.recorder = &MockProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProcessor) EXPECT() *MockProcessorMockRecorder {
	return m.recorder
}

// AggregatorFor mocks base method.
func (m *MockProcessor) AggregatorFor(descriptor *sdkapi.Descriptor, aggregator ...*aggregator.Aggregator) {
	m.ctrl.T.Helper()
	varargs := []interface{}{descriptor}
	for _, a := range aggregator {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AggregatorFor", varargs...)
}

// AggregatorFor indicates an expected call of AggregatorFor.
func (mr *MockProcessorMockRecorder) AggregatorFor(descriptor interface{}, aggregator ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{descriptor}, aggregator...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AggregatorFor", reflect.TypeOf((*MockProcessor)(nil).AggregatorFor), varargs...)
}

// Process mocks base method.
func (m *MockProcessor) Process(accum export.Accumulation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Process", accum)
	ret0, _ := ret[0].(error)
	return ret0
}

// Process indicates an expected call of Process.
func (mr *MockProcessorMockRecorder) Process(accum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockProcessor)(nil).Process), accum)
}

// MockAggregatorSelector is a mock of AggregatorSelector interface.
type MockAggregatorSelector struct {
	ctrl     *gomock.Controller
	recorder *MockAggregatorSelectorMockRecorder
}

// MockAggregatorSelectorMockRecorder is the mock recorder for MockAggregatorSelector.
type MockAggregatorSelectorMockRecorder struct {
	mock *MockAggregatorSelector
}

// NewMockAggregatorSelector creates a new mock instance.
func NewMockAggregatorSelector(ctrl *gomock.Controller) *MockAggregatorSelector {
	mock := &MockAggregatorSelector{ctrl: ctrl}
	mock.recorder = &MockAggregatorSelectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAggregatorSelector) EXPECT() *MockAggregatorSelectorMockRecorder {
	return m.recorder
}

// AggregatorFor mocks base method.
func (m *MockAggregatorSelector) AggregatorFor(descriptor *sdkapi.Descriptor, aggregator ...*aggregator.Aggregator) {
	m.ctrl.T.Helper()
	varargs := []interface{}{descriptor}
	for _, a := range aggregator {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AggregatorFor", varargs...)
}

// AggregatorFor indicates an expected call of AggregatorFor.
func (mr *MockAggregatorSelectorMockRecorder) AggregatorFor(descriptor interface{}, aggregator ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{descriptor}, aggregator...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AggregatorFor", reflect.TypeOf((*MockAggregatorSelector)(nil).AggregatorFor), varargs...)
}

// MockCheckpointer is a mock of Checkpointer interface.
type MockCheckpointer struct {
	ctrl     *gomock.Controller
	recorder *MockCheckpointerMockRecorder
}

// MockCheckpointerMockRecorder is the mock recorder for MockCheckpointer.
type MockCheckpointerMockRecorder struct {
	mock *MockCheckpointer
}

// NewMockCheckpointer creates a new mock instance.
func NewMockCheckpointer(ctrl *gomock.Controller) *MockCheckpointer {
	mock := &MockCheckpointer{ctrl: ctrl}
	mock.recorder = &MockCheckpointerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCheckpointer) EXPECT() *MockCheckpointerMockRecorder {
	return m.recorder
}

// AggregatorFor mocks base method.
func (m *MockCheckpointer) AggregatorFor(descriptor *sdkapi.Descriptor, aggregator ...*aggregator.Aggregator) {
	m.ctrl.T.Helper()
	varargs := []interface{}{descriptor}
	for _, a := range aggregator {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AggregatorFor", varargs...)
}

// AggregatorFor indicates an expected call of AggregatorFor.
func (mr *MockCheckpointerMockRecorder) AggregatorFor(descriptor interface{}, aggregator ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{descriptor}, aggregator...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AggregatorFor", reflect.TypeOf((*MockCheckpointer)(nil).AggregatorFor), varargs...)
}

// FinishCollection mocks base method.
func (m *MockCheckpointer) FinishCollection() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FinishCollection")
	ret0, _ := ret[0].(error)
	return ret0
}

// FinishCollection indicates an expected call of FinishCollection.
func (mr *MockCheckpointerMockRecorder) FinishCollection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FinishCollection", reflect.TypeOf((*MockCheckpointer)(nil).FinishCollection))
}

// Process mocks base method.
func (m *MockCheckpointer) Process(accum export.Accumulation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Process", accum)
	ret0, _ := ret[0].(error)
	return ret0
}

// Process indicates an expected call of Process.
func (mr *MockCheckpointerMockRecorder) Process(accum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockCheckpointer)(nil).Process), accum)
}

// Reader mocks base method.
func (m *MockCheckpointer) Reader() export.Reader {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reader")
	ret0, _ := ret[0].(export.Reader)
	return ret0
}

// Reader indicates an expected call of Reader.
func (mr *MockCheckpointerMockRecorder) Reader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reader", reflect.TypeOf((*MockCheckpointer)(nil).Reader))
}

// StartCollection mocks base method.
func (m *MockCheckpointer) StartCollection() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartCollection")
}

// StartCollection indicates an expected call of StartCollection.
func (mr *MockCheckpointerMockRecorder) StartCollection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCollection", reflect.TypeOf((*MockCheckpointer)(nil).StartCollection))
}

// MockCheckpointerFactory is a mock of CheckpointerFactory interface.
type MockCheckpointerFactory struct {
	ctrl     *gomock.Controller
	recorder *MockCheckpointerFactoryMockRecorder
}

// MockCheckpointerFactoryMockRecorder is the mock recorder for MockCheckpointerFactory.
type MockCheckpointerFactoryMockRecorder struct {
	mock *MockCheckpointerFactory
}

// NewMockCheckpointerFactory creates a new mock instance.
func NewMockCheckpointerFactory(ctrl *gomock.Controller) *MockCheckpointerFactory {
	mock := &MockCheckpointerFactory{ctrl: ctrl}
	mock.recorder = &MockCheckpointerFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCheckpointerFactory) EXPECT() *MockCheckpointerFactoryMockRecorder {
	return m.recorder
}

// NewCheckpointer mocks base method.
func (m *MockCheckpointerFactory) NewCheckpointer() export.Checkpointer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewCheckpointer")
	ret0, _ := ret[0].(export.Checkpointer)
	return ret0
}

// NewCheckpointer indicates an expected call of NewCheckpointer.
func (mr *MockCheckpointerFactoryMockRecorder) NewCheckpointer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewCheckpointer", reflect.TypeOf((*MockCheckpointerFactory)(nil).NewCheckpointer))
}

// MockExporter is a mock of Exporter interface.
type MockExporter struct {
	ctrl     *gomock.Controller
	recorder *MockExporterMockRecorder
}

// MockExporterMockRecorder is the mock recorder for MockExporter.
type MockExporterMockRecorder struct {
	mock *MockExporter
}

// NewMockExporter creates a new mock instance.
func NewMockExporter(ctrl *gomock.Controller) *MockExporter {
	mock := &MockExporter{ctrl: ctrl}
	mock.recorder = &MockExporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExporter) EXPECT() *MockExporterMockRecorder {
	return m.recorder
}

// Export mocks base method.
func (m *MockExporter) Export(ctx context.Context, resource *resource.Resource, reader export.InstrumentationLibraryReader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Export", ctx, resource, reader)
	ret0, _ := ret[0].(error)
	return ret0
}

// Export indicates an expected call of Export.
func (mr *MockExporterMockRecorder) Export(ctx, resource, reader interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Export", reflect.TypeOf((*MockExporter)(nil).Export), ctx, resource, reader)
}

// TemporalityFor mocks base method.
func (m *MockExporter) TemporalityFor(descriptor *sdkapi.Descriptor, aggregationKind aggregation.Kind) aggregation.Temporality {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TemporalityFor", descriptor, aggregationKind)
	ret0, _ := ret[0].(aggregation.Temporality)
	return ret0
}

// TemporalityFor indicates an expected call of TemporalityFor.
func (mr *MockExporterMockRecorder) TemporalityFor(descriptor, aggregationKind interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TemporalityFor", reflect.TypeOf((*MockExporter)(nil).TemporalityFor), descriptor, aggregationKind)
}

// MockInstrumentationLibraryReader is a mock of InstrumentationLibraryReader interface.
type MockInstrumentationLibraryReader struct {
	ctrl     *gomock.Controller
	recorder *MockInstrumentationLibraryReaderMockRecorder
}

// MockInstrumentationLibraryReaderMockRecorder is the mock recorder for MockInstrumentationLibraryReader.
type MockInstrumentationLibraryReaderMockRecorder struct {
	mock *MockInstrumentationLibraryReader
}

// NewMockInstrumentationLibraryReader creates a new mock instance.
func NewMockInstrumentationLibraryReader(ctrl *gomock.Controller) *MockInstrumentationLibraryReader {
	mock := &MockInstrumentationLibraryReader{ctrl: ctrl}
	mock.recorder = &MockInstrumentationLibraryReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInstrumentationLibraryReader) EXPECT() *MockInstrumentationLibraryReaderMockRecorder {
	return m.recorder
}

// ForEach mocks base method.
func (m *MockInstrumentationLibraryReader) ForEach(readerFunc func(instrumentation.Library, export.Reader) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForEach", readerFunc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForEach indicates an expected call of ForEach.
func (mr *MockInstrumentationLibraryReaderMockRecorder) ForEach(readerFunc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForEach", reflect.TypeOf((*MockInstrumentationLibraryReader)(nil).ForEach), readerFunc)
}

// MockReader is a mock of Reader interface.
type MockReader struct {
	ctrl     *gomock.Controller
	recorder *MockReaderMockRecorder
}

// MockReaderMockRecorder is the mock recorder for MockReader.
type MockReaderMockRecorder struct {
	mock *MockReader
}

// NewMockReader creates a new mock instance.
func NewMockReader(ctrl *gomock.Controller) *MockReader {
	mock := &MockReader{ctrl: ctrl}
	mock.recorder = &MockReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReader) EXPECT() *MockReaderMockRecorder {
	return m.recorder
}

// ForEach mocks base method.
func (m *MockReader) ForEach(tempSelector aggregation.TemporalitySelector, recordFunc func(export.Record) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForEach", tempSelector, recordFunc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ForEach indicates an expected call of ForEach.
func (mr *MockReaderMockRecorder) ForEach(tempSelector, recordFunc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForEach", reflect.TypeOf((*MockReader)(nil).ForEach), tempSelector, recordFunc)
}

// Lock mocks base method.
func (m *MockReader) Lock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Lock")
}

// Lock indicates an expected call of Lock.
func (mr *MockReaderMockRecorder) Lock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockReader)(nil).Lock))
}

// RLock mocks base method.
func (m *MockReader) RLock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RLock")
}

// RLock indicates an expected call of RLock.
func (mr *MockReaderMockRecorder) RLock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RLock", reflect.TypeOf((*MockReader)(nil).RLock))
}

// RUnlock mocks base method.
func (m *MockReader) RUnlock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RUnlock")
}

// RUnlock indicates an expected call of RUnlock.
func (mr *MockReaderMockRecorder) RUnlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RUnlock", reflect.TypeOf((*MockReader)(nil).RUnlock))
}

// Unlock mocks base method.
func (m *MockReader) Unlock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unlock")
}

// Unlock indicates an expected call of Unlock.
func (mr *MockReaderMockRecorder) Unlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockReader)(nil).Unlock))
}