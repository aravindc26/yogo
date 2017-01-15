package yoga

import (
	"errors"
	"math"
)

type Size struct {
	width  float64
	height float64
}

type Value struct {
	value float64
	unit  Unit
}

type CachedMeasurement struct {
	avaialbleWidth    float64
	availableHeight   float64
	widthMeasureMode  MeasureMode
	heightMeasureMode MeasureMode
	computedWidth     float64
	computedHeight    float64
}

type Layout struct {
	position                    [4]float64
	dimensions                  [2]float64
	padding                     [6]float64
	direction                   Direction
	computedFlexBasisGeneration uint32
	computedFlexBasis           float64
	generationCount             uint32
	lastParentDirection         Direction
	nextCachedMeasurementsIndex uint32
	cachedMeasurements          [16]CachedMeasurement
	measuredDimensions          [2]float64
	cachedLayout                CachedMeasurement
}

type Style struct {
	direction      Direction
	flexDirection  FlexDirection
	justifyContent Justify
	alignContent   Align
	alignItems     Align
	alignSelf      Align
	positionType   PositionType
	flexWrap       Wrap
	overflow       Overflow
	flex           float64
	flexGrow       float64
	flexShrink     float64
	flexBasis      Value
	margin         [9]Value
	position       [9]Value
	padding        [9]Value
	border         [9]Value
	dimensions     [2]Value
	minDimensions  [2]Value
	maxDimensions  [2]Value
	aspectRatio    float64
}

type MeasureFunc func(node *Node, width float64, widthMode MeasureMode, height float64, heightMode MeasureMode) Size
type BaseLineFunc func(node *Node, width, height float64) float64
type PrintFunc func(node *Node)

type Node struct {
	style        Style
	layout       Layout
	lineIndex    uint32
	parent       *Node
	children     []*Node
	nextChild    *Node
	measure      MeasureFunc
	baseLine     BaseLineFunc
	print        PrintFunc
	context      *interface{}
	isDirty      bool
	hasNewLayout bool
}

func ComputedEdgeValue(edges [9]Value, edge Edge, defaultValue *Value) (*Value, error) {
	if !(edge <= EdgeEnd) {
		return nil, errors.New("Cannot get computed value of multi-edge shorthands")
	}
	if edges[edge].unit != UnitUndefined {
		return &edges[edge], nil
	}
	if (edge == EdgeTop || edge == EdgeBottom) && edges[EdgeVertical].unit != UnitUndefined {
		return &edges[EdgeVertical], nil
	}
	if (edge == EdgeLeft || edge == EdgeRight || edge == EdgeStart || edge == EdgeEnd) &&
		edges[EdgeHorizontal].unit != UnitUndefined {
		return &edges[EdgeHorizontal], nil
	}
	if edges[EdgeAll].unit != UnitUndefined {
		return &edges[EdgeAll], nil
	}
	if edge == EdgeStart || edge == EdgeEnd {
		return &Value{value: math.NaN()}, nil
	}
	return defaultValue, nil
}

func ValueResolve(unit *Value, parentSize float64) float64 {
	if unit.unit == UnitPixel {
		return unit.value
	}
	return unit.value * parentSize / 100.0
}

func GetChildCount(node *Node) int {
	return len(node.children)
}

func MarkDirtyInternal(node *Node) {
	if !node.isDirty {
		node.isDirty = true
		node.layout.computedFlexBasis = math.NaN()
		if node.parent != nil {
			MarkDirtyInternal(node.parent)
		}
	}
}
func SetMeasureFunc(node *Node, measureFunc MeasureFunc) error {
	if measureFunc == nil {
		node.measure = nil
		return nil
	}
	if GetChildCount(node) != 0 {
		return errors.New("Cannot set measure function: Nodes with measure functions cannot have children.")
	}
	node.measure = measureFunc
	return nil
}

func GetMeasureFunc(node *Node) MeasureFunc {
	return node.measure
}

func InsertChild(node *Node, child *Node, index int) error {
	if child.parent != nil {
		return errors.New("Child already has parent, it must be removed first.")
	}
	if node.measure != nil {
		return errors.New("Cannot add child: Nodes with measure functions cannot have children")
	}
	node.children = append(node.children[:index], append([]*Node{child}, node.children[index:]...)...)
	child.parent = node
	MarkDirtyInternal(node)
	return nil
}

func RemoveChild(node *Node, child *Node) {
	if listDelete(node.children, child) {
		child.parent = nil
		MarkDirtyInternal(node)
	}
}

func listDelete(nodes []*Node, item *Node) bool {
	for i := 0; i < len(nodes); i++ {
		if nodes[i] == item {
			copy(nodes[i:], nodes[i+1:])
			nodes[len(nodes)-1] = nil
			nodes = nodes[:len(nodes)-1]
			return true
		}
	}
	return false
}

func GetChild(node *Node, index int) *Node {
	return node.children[index]
}

func GetParent(node *Node) *Node {
	return node.parent
}

func MarkDirty(node *Node) error {
	if node.measure == nil {
		return errors.New("Only leaf nodes with custom measure functions should manually mark themselves dirty")
	}
	MarkDirtyInternal(node)
	return nil
}

func IsDirty(node *Node) bool {
	return node.isDirty
}

func CopyStyle(dstNode, srcNode *Node) {
	dstNode.style = srcNode.style
	MarkDirtyInternal(dstNode)
}

func GetFlexGrow(node *Node) float64 {
	if !math.IsNaN(node.style.flexGrow) {
		return node.style.flexGrow
	}
	if !math.IsNaN(node.style.flex) && node.style.flex > 0.0 {
		return node.style.flex
	}
	return 0.0
}

func GetFlexShrink(node *Node) float64 {
	if !math.IsNaN(node.style.flexShrink) {
		return node.style.flexShrink
	}
	if !math.IsNaN(node.style.flex) && node.style.flex < 0.0 {
		return -node.style.flex
	}
	return 0.0
}

func GetFlexBasisPtr(node *Node) *Value {
	if node.style.flexBasis.unit != UnitUndefined {
		return &node.style.flexBasis
	}
	if !math.IsNaN(node.style.flex) && node.style.flex > 0.0 {
		return &Value{value: 0.0, unit: UnitPixel}
	}
	return &Value{value: math.NaN()}
}

func GetFlexBasis(node *Node) Value {
	return *GetFlexBasisPtr(node)
}

func SetFlex(node *Node, flex float64) {
	if node.style.flex != flex {
		node.style.flex = flex
		MarkDirtyInternal(node)
	}
}

func SetContext(node *Node, context *interface{}) {
	node.context = context
}

func NodeGetContext(node *Node) *interface{} {
	return node.context
}

func SetPrintFunc(node *Node, printFunc PrintFunc) {
	node.print = printFunc
}

func GetPrintFunc(node *Node) PrintFunc {
	return node.print
}

func SetHasNewLayout(node *Node, hasNewLayout bool) {
	node.hasNewLayout = hasNewLayout
}

func GetHasNewLayout(node *Node) bool {
	return node.hasNewLayout
}

func SetDirection(node *Node, direction Direction) {
	if node.style.direction != direction {
		node.style.direction = direction
		MarkDirtyInternal(node)
	}
}

func GetDirection(node *Node) Direction {
	return node.style.direction
}

func SetFlexDirection(node *Node, flexDirection FlexDirection) {
	if node.style.flexDirection != flexDirection {
		node.style.flexDirection = flexDirection
		MarkDirtyInternal(node)
	}
}

func GetFlexDirection(node *Node) FlexDirection {
	return node.style.flexDirection
}

func SetJustifyContent(node *Node, justifyContent Justify) {
	if node.style.justifyContent != justifyContent {
		node.style.justifyContent = justifyContent
		MarkDirtyInternal(node)
	}
}

func GetJustifyContent(node *Node) Justify {
	return node.style.justifyContent
}

func SetAlignContent(node *Node, alignContent Align) {
	if node.style.alignContent != alignContent {
		node.style.alignContent = alignContent
		MarkDirtyInternal(node)
	}
}

func GetAlignContent(node *Node) Align {
	return node.style.alignContent
}

func SetAlignItems(node *Node, alignItems Align) {
	if node.style.alignItems != alignItems {
		node.style.alignItems = alignItems
		MarkDirtyInternal(node)
	}
}

func GetAlignItems(node *Node) Align {
	return node.style.alignItems
}

func SetAlignSelf(node *Node, alignSelf Align) {
	if node.style.alignSelf != alignSelf {
		node.style.alignSelf = alignSelf
		MarkDirtyInternal(node)
	}
}

func GetAlignSelf(node *Node) Align {
	return node.style.alignSelf
}

func SetPositionType(node *Node, positionType PositionType) {
	if node.style.positionType != positionType {
		node.style.positionType = positionType
		MarkDirtyInternal(node)
	}
}

func GetPositionType(node *Node) PositionType {
	return node.style.positionType
}

func SetFlexWrap(node *Node, flexWrap Wrap) {
	if node.style.flexWrap != flexWrap {
		node.style.flexWrap = flexWrap
		MarkDirtyInternal(node)
	}
}

func GetFlexWrap(node *Node) Wrap {
	return node.style.flexWrap
}

func SetOverflow(node *Node, overflow Overflow) {
	if node.style.overflow != overflow {
		node.style.overflow = overflow
		MarkDirtyInternal(node)
	}
}

func GetOverflow(node *Node) Overflow {
	return node.style.overflow
}

func SetFlexGrow(node *Node, flexGrow float64) {
	if node.style.flexGrow != flexGrow {
		node.style.flexGrow = flexGrow
		MarkDirtyInternal(node)
	}
}

func SetFlexShrink(node *Node, flexShrink float64) {
	if node.style.flexShrink != flexShrink {
		node.style.flexShrink = flexShrink
		MarkDirtyInternal(node)
	}
}

func SetFlexBasis(node *Node, flexBasis float64) {
	if node.style.flexBasis.value != flexBasis || node.style.flexBasis.unit != UnitPixel {
		node.style.flexBasis.value = flexBasis
		if !math.IsNaN(flexBasis) {
			node.style.flexBasis.unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetFlexBasisPercent(node *Node, flexBasis float64) {
	if node.style.flexBasis.value != flexBasis || node.style.flexBasis.unit != UnitPercent {
		node.style.flexBasis.value = flexBasis
		if !math.IsNaN(flexBasis) {
			node.style.flexBasis.unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func SetPosition(node *Node, edge Edge, position float64) {
	if node.style.position[int(edge)].value != position || node.style.position[int(edge)].unit != UnitPixel {
		node.style.position[int(edge)].value = position
		if !math.IsNaN(position) {
			node.style.position[int(edge)].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetPositionPercent(node *Node, edge Edge, position float64) {
	if node.style.position[int(edge)].value != position || node.style.position[int(edge)].unit != UnitPercent {
		node.style.position[int(edge)].value = position
		if !math.IsNaN(position) {
			node.style.position[int(edge)].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetPosition(node *Node, edge Edge) (Value, error) {
	r, err := ComputedEdgeValue(node.style.position, edge, &Value{value: math.NaN()})
	if err != nil {
		return Value{}, err
	}
	return *r, nil
}

func SetMargin(node *Node, edge Edge, margin float64) {
	if node.style.margin[int(edge)].value != margin || node.style.margin[int(edge)].unit != UnitPixel {
		node.style.margin[int(edge)].value = margin
		if !math.IsNaN(margin) {
			node.style.margin[int(edge)].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetMarginPercent(node *Node, edge Edge, margin float64) {
	if node.style.margin[int(edge)].value != margin || node.style.margin[int(edge)].unit != UnitPercent {
		node.style.margin[int(edge)].value = margin
		if !math.IsNaN(margin) {
			node.style.margin[int(edge)].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetMargin(node *Node, edge Edge) (Value, error) {
	r, err := ComputedEdgeValue(node.style.margin, edge, &Value{unit: UnitPixel})
	if err != nil {
		return Value{}, err
	}
	return *r, nil
}

func SetPadding(node *Node, edge Edge, padding float64) {
	if node.style.padding[int(edge)].value != padding || node.style.padding[int(edge)].unit != UnitPixel {
		node.style.padding[int(edge)].value = padding
		if !math.IsNaN(padding) {
			node.style.padding[int(edge)].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetPaddingPercent(node *Node, edge Edge, padding float64) {
	if node.style.padding[int(edge)].value != padding || node.style.padding[int(edge)].unit != UnitPercent {
		node.style.padding[int(edge)].value = padding
		if !math.IsNaN(padding) {
			node.style.padding[int(edge)].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetPadding(node *Node, edge Edge) (Value, error) {
	r, err := ComputedEdgeValue(node.style.padding, edge, &Value{unit: UnitPixel})
	if err != nil {
		return Value{}, err
	}
	return *r, nil
}

func SetBorder(node *Node, edge Edge, border float64) {
	if node.style.border[int(edge)].value != border || node.style.border[int(edge)].unit != UnitPixel {
		node.style.border[int(edge)].value = border
		if !math.IsNaN(border) {
			node.style.border[int(edge)].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func GetBorder(node *Node, edge Edge) (float64, error) {
	r, err := ComputedEdgeValue(node.style.border, edge, &Value{unit: UnitPixel})
	if err != nil {
		return math.NaN(), err
	}
	return r.value, nil
}
