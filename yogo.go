package yoga

import (
	"errors"
	"log"
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
	margin                      [6]float64
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

func SetWidth(node *Node, width float64) {
	if node.style.dimensions[DimensionWidth].value != width || node.style.dimensions[DimensionWidth].unit != UnitPixel {
		node.style.dimensions[DimensionWidth].value = width
		if !math.IsNaN(width) {
			node.style.dimensions[DimensionWidth].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetWidthPercent(node *Node, width float64) {
	if node.style.dimensions[DimensionWidth].value != width || node.style.dimensions[DimensionWidth].unit != UnitPercent {
		node.style.dimensions[DimensionWidth].value = width
		if !math.IsNaN(width) {
			node.style.dimensions[DimensionWidth].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetStyleWidth(node *Node) Value {
	return node.style.dimensions[DimensionWidth]
}

func SetHeight(node *Node, height float64) {
	if node.style.dimensions[DimensionHeight].value != height || node.style.dimensions[DimensionWidth].unit != UnitPixel {
		node.style.dimensions[DimensionHeight].value = height
		if !math.IsNaN(height) {
			node.style.dimensions[DimensionHeight].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetHeightPercent(node *Node, height float64) {
	if node.style.dimensions[DimensionHeight].value != height || node.style.dimensions[DimensionHeight].unit != UnitPercent {
		node.style.dimensions[DimensionHeight].value = height
		if !math.IsNaN(height) {
			node.style.dimensions[DimensionHeight].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetHeight(node *Node) Value {
	return node.style.dimensions[DimensionHeight]
}

func SetMinWidth(node *Node, minWidth float64) {
	if node.style.minDimensions[DimensionWidth].value != minWidth || node.style.minDimensions[DimensionWidth].unit != UnitPixel {
		node.style.minDimensions[DimensionWidth].value = minWidth
		if !math.IsNaN(minWidth) {
			node.style.minDimensions[DimensionWidth].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetMinWidthPercent(node *Node, minWidth float64) {
	if node.style.minDimensions[DimensionWidth].value != minWidth || node.style.minDimensions[DimensionWidth].unit != UnitPercent {
		node.style.minDimensions[DimensionWidth].value = minWidth
		if !math.IsNaN(minWidth) {
			node.style.minDimensions[DimensionWidth].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetMinWidth(node *Node) Value {
	return node.style.minDimensions[DimensionWidth]
}

func SetMinHeight(node *Node, minHeight float64) {
	if node.style.minDimensions[DimensionHeight].value != minHeight || node.style.minDimensions[DimensionHeight].unit != UnitPixel {
		node.style.minDimensions[DimensionHeight].value = minHeight
		if !math.IsNaN(minHeight) {
			node.style.minDimensions[DimensionHeight].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetMinHeightPercent(node *Node, minHeight float64) {
	if node.style.minDimensions[DimensionHeight].value != minHeight || node.style.minDimensions[DimensionHeight].unit != UnitPercent {
		node.style.minDimensions[DimensionHeight].value = minHeight
		if !math.IsNaN(minHeight) {
			node.style.minDimensions[DimensionHeight].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetMinHeight(node *Node) Value {
	return node.style.minDimensions[DimensionHeight]
}

func SetMaxWidth(node *Node, maxWidth float64) {
	if node.style.maxDimensions[DimensionWidth].value != maxWidth || node.style.maxDimensions[DimensionWidth].unit != UnitPixel {
		node.style.maxDimensions[DimensionWidth].value = maxWidth
		if !math.IsNaN(maxWidth) {
			node.style.maxDimensions[DimensionWidth].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetMaxWidthPercent(node *Node, maxWidth float64) {
	if node.style.maxDimensions[DimensionWidth].value != maxWidth || node.style.maxDimensions[DimensionWidth].unit != UnitPercent {
		node.style.maxDimensions[DimensionWidth].value = maxWidth
		if !math.IsNaN(maxWidth) {
			node.style.maxDimensions[DimensionWidth].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetMaxWidth(node *Node) Value {
	return node.style.maxDimensions[DimensionWidth]
}

func SetMaxHeight(node *Node, maxHeight float64) {
	if node.style.maxDimensions[DimensionHeight].value != maxHeight || node.style.maxDimensions[DimensionHeight].unit != UnitPixel {
		node.style.maxDimensions[DimensionHeight].value = maxHeight
		if !math.IsNaN(maxHeight) {
			node.style.maxDimensions[DimensionHeight].unit = UnitPixel
		}
		MarkDirtyInternal(node)
	}
}

func SetMaxHeightPercent(node *Node, maxHeight float64) {
	if node.style.maxDimensions[DimensionHeight].value != maxHeight || node.style.maxDimensions[DimensionHeight].unit != UnitPercent {
		node.style.maxDimensions[DimensionHeight].value = maxHeight
		if !math.IsNaN(maxHeight) {
			node.style.maxDimensions[DimensionHeight].unit = UnitPercent
		}
		MarkDirtyInternal(node)
	}
}

func GetMaxHeight(node *Node) Value {
	return node.style.maxDimensions[DimensionHeight]
}

func SetAspectRatio(node *Node, aspectRatio float64) {
	if node.style.aspectRatio != aspectRatio {
		node.style.aspectRatio = aspectRatio
		MarkDirtyInternal(node)
	}
}

func GetLayoutAspectRatio(node *Node) float64 {
	return node.style.aspectRatio
}

func GetLayoutLeft(node *Node) float64 {
	return node.layout.position[EdgeLeft]
}

func GetLayoutTop(node *Node) float64 {
	return node.layout.position[EdgeTop]
}

func GetLayoutRight(node *Node) float64 {
	return node.layout.position[EdgeRight]
}

func GetLayoutBottom(node *Node) float64 {
	return node.layout.position[EdgeBottom]
}

func GetLayoutWidth(node *Node) float64 {
	return node.layout.dimensions[DimensionWidth]
}

func GetLayoutHeight(node *Node) float64 {
	return node.layout.dimensions[DimensionHeight]
}

func GetLayoutDirection(node *Node) Direction {
	return node.layout.direction
}

func GetLayoutMargin(node *Node, edge Edge) (float64, error) {
	if !(edge <= EdgeEnd) {
		return 0.0, errors.New("Cannot get layout properties of multi-edge shorthands")
	}
	if edge == EdgeLeft {
		if node.layout.direction == DirectionRTL {
			return node.layout.margin[EdgeEnd], nil
		} else {
			return node.layout.margin[EdgeStart], nil
		}
	}
	if edge == EdgeRight {
		if node.layout.direction == DirectionRTL {
			return node.layout.margin[EdgeStart], nil
		} else {
			return node.layout.margin[EdgeEnd], nil
		}
	}
	return node.layout.margin[edge], nil
}

func GetLayoutPadding(node *Node, edge Edge) (float64, error) {
	if !(edge <= EdgeEnd) {
		return 0.0, errors.New("Cannot get layout properties of multi-edge shorthands")
	}
	if edge == EdgeLeft {
		if node.layout.direction == DirectionRTL {
			return node.layout.padding[EdgeEnd], nil
		} else {
			return node.layout.padding[EdgeStart], nil
		}
	}
	if edge == EdgeRight {
		if node.layout.direction == DirectionRTL {
			return node.layout.padding[EdgeStart], nil
		} else {
			return node.layout.padding[EdgeEnd], nil
		}
	}
	return node.layout.padding[edge], nil
}

var currentGenerationCount uint32

func ValueEqual(a Value, b Value) bool {
	if a.unit != b.unit {
		return false
	}
	if a.unit == UnitUndefined {
		return true
	}
	return math.Abs(a.value-b.value) < 0.0001
}

func FloatsEqual(a, b float64) bool {
	if math.IsNaN(a) {
		return math.IsNaN(b)
	}
	return math.Abs(a-b) < 0.0001
}

func Indent(n int) {
	for i := 0; i < n; i++ {
		log.Println("  ")
	}
}

func PrintNumberIfNotZero(str string, number *Value) {
	if !FloatsEqual(number.value, 0) {
		log.Printf("%s: %g%s, ", str, number.value, number.unit)
	}
}

func PrintNumberIfNotUndefinedf(str string, number float64) {
	if !math.IsNaN(number) {
		log.Printf("%s: %g, ", str, number)
	}
}

func PrintNumberIfNotUndefined(str string, number *Value) {
	if number.unit != UnitUndefined {
		log.Printf("%s: %g%s, ", str, number.value, number.unit)
	}
}

func FourValuesEqual(four [4]Value) bool {
	return ValueEqual(four[0], four[1]) && ValueEqual(four[0], four[2]) && ValueEqual(four[0], four[3])
}

func NodePrintInternal(node *Node, options PrintOptions, level int) error {
	Indent(level)
	log.Print("{")
	if node.print != nil {
		node.print(node)
	}
	if options&PrintOptionsLayout != 0 {
		log.Print("layout: {")
		log.Print("width: %g, ", node.layout.dimensions[DimensionWidth])
		log.Print("height: %g, ", node.layout.dimensions[DimensionHeight])
		log.Print("top: %g, ", node.layout.position[EdgeTop])
		log.Print("left: %g", node.layout.position[EdgeLeft])
		log.Print("}, ")
	}

	if options&PrintOptionsStyle != 0 {
		log.Printf("flexDirection: '%s', ", node.style.flexDirection)
		log.Printf("justifyContent: '%s', ", node.style.justifyContent)
		log.Printf("alignItems: '%s', ", node.style.alignItems)
		log.Printf("alignContent: '%s', ", node.style.alignContent)
		log.Printf("alignSelf: '%s', ", node.style.alignSelf)
		PrintNumberIfNotUndefinedf("flexGrow", GetFlexGrow(node))
		PrintNumberIfNotUndefinedf("flexShrink", GetFlexShrink(node))
		PrintNumberIfNotUndefined("flexBasis", GetFlexBasisPtr(node))
		log.Printf("overflow: '%s', ", node.style.overflow)
		var fourVal [4]Value
		for i := 0; i < 4; i++ {
			fourVal[i] = node.style.margin[i]
		}
		if FourValuesEqual(fourVal) {
			val, err := ComputedEdgeValue(node.style.margin, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("margin", val)
		} else {
			val, err := ComputedEdgeValue(node.style.margin, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginLeft", val)
			val, err = ComputedEdgeValue(node.style.margin, EdgeRight, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginRight", val)
			val, err = ComputedEdgeValue(node.style.margin, EdgeTop, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginTop", val)
			val, err = ComputedEdgeValue(node.style.margin, EdgeBottom, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginBottom", val)
			val, err = ComputedEdgeValue(node.style.margin, EdgeStart, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginStart", val)
			val, err = ComputedEdgeValue(node.style.margin, EdgeEnd, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("marginEnd", val)
		}
		for i := 0; i < 4; i++ {
			fourVal[i] = node.style.border[i]
		}
		if FourValuesEqual(fourVal) {
			val, err := ComputedEdgeValue(node.style.border, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderWidth", val)
		} else {
			val, err := ComputedEdgeValue(node.style.border, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderLeftWidth", val)
			val, err = ComputedEdgeValue(node.style.border, EdgeRight, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderRightWidth", val)
			val, err = ComputedEdgeValue(node.style.border, EdgeTop, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderTopWidth", val)
			val, err = ComputedEdgeValue(node.style.border, EdgeBottom, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderBottomWidth", val)
			val, err = ComputedEdgeValue(node.style.border, EdgeStart, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderStartWidth", val)
			val, err = ComputedEdgeValue(node.style.border, EdgeEnd, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("borderEndWidth", val)
		}
		for i := 0; i < 4; i++ {
			fourVal[i] = node.style.padding[i]
		}
		if FourValuesEqual(fourVal) {
			val, err := ComputedEdgeValue(node.style.padding, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("padding", val)
		} else {
			val, err := ComputedEdgeValue(node.style.padding, EdgeLeft, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingLeft", val)
			val, err = ComputedEdgeValue(node.style.padding, EdgeRight, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingRight", val)
			val, err = ComputedEdgeValue(node.style.padding, EdgeTop, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingTop", val)
			val, err = ComputedEdgeValue(node.style.padding, EdgeBottom, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingBottom", val)
			val, err = ComputedEdgeValue(node.style.padding, EdgeStart, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingStart", val)
			val, err = ComputedEdgeValue(node.style.padding, EdgeEnd, &Value{value: 0, unit: UnitPixel})
			if err != nil {
				return err
			}
			PrintNumberIfNotZero("paddingEnd", val)
		}
		PrintNumberIfNotUndefined("width", &node.style.dimensions[DimensionWidth])
		PrintNumberIfNotUndefined("height", &node.style.dimensions[DimensionHeight])
		PrintNumberIfNotUndefined("maxWidth", &node.style.maxDimensions[DimensionWidth])
		PrintNumberIfNotUndefined("maxHeigth", &node.style.maxDimensions[DimensionHeight])
		PrintNumberIfNotUndefined("minWidth", &node.style.minDimensions[DimensionWidth])
		PrintNumberIfNotUndefined("minHeight", &node.style.minDimensions[DimensionHeight])
		log.Printf("position: '%s', ", node.style.positionType)
		val, err := ComputedEdgeValue(node.style.position, EdgeLeft, &Value{value: math.NaN(), unit: UnitUndefined})
		if err != nil {
			return err
		}
		PrintNumberIfNotUndefined("left", val)
		val, err = ComputedEdgeValue(node.style.position, EdgeRight, &Value{value: math.NaN(), unit: UnitUndefined})
		if err != nil {
			return err
		}
		PrintNumberIfNotUndefined("right", val)
		val, err = ComputedEdgeValue(node.style.position, EdgeTop, &Value{value: math.NaN(), unit: UnitUndefined})
		if err != nil {
			return err
		}
		PrintNumberIfNotUndefined("top", val)
		val, err = ComputedEdgeValue(node.style.position, EdgeBottom, &Value{value: math.NaN(), unit: UnitUndefined})
		if err != nil {
			return err
		}
		PrintNumberIfNotUndefined("bottom", val)

	}
	childCount := len(node.children)
	if (options&PrintOptionsChildren != 0) && childCount > 0 {
		log.Print("children: [\n")
		for i := 0; i < childCount; i++ {
			err := NodePrintInternal(GetChild(node, i), options, level+1)
			if err != nil {
				return err
			}
		}
		Indent(level)
		log.Print("]},\n")
	} else {
		log.Print("]},\n")
	}
	return nil
}

func NodePrint(node *Node, options PrintOptions) error {
	return NodePrintInternal(node, options, 0)
}

var leading [4]Edge

func init() {
	leading[FlexDirectionColumn] = EdgeTop
	leading[FlexDirectionColumnReverse] = EdgeBottom
	leading[FlexDirectionRow] = EdgeLeft
	leading[FlexDirectionRowReverse] = EdgeRight
}

var trailing [4]Edge

func init() {
	trailing[FlexDirectionColumn] = EdgeBottom
	trailing[FlexDirectionColumnReverse] = EdgeTop
	trailing[FlexDirectionRow] = EdgeRight
	trailing[FlexDirectionRowReverse] = EdgeLeft
}

var pos [4]Edge

func init() {
	pos[FlexDirectionColumn] = EdgeTop
	pos[FlexDirectionColumnReverse] = EdgeBottom
	pos[FlexDirectionRow] = EdgeLeft
	pos[FlexDirectionRowReverse] = EdgeRight
}

var dim [4]Dimension

func init() {
	dim[FlexDirectionColumn] = DimensionHeight
	dim[FlexDirectionColumnReverse] = DimensionHeight
	dim[FlexDirectionRow] = DimensionWidth
	dim[FlexDirectionRowReverse] = DimensionWidth
}

func FlexDirectionIsRow(direction FlexDirection) bool {
	return direction == FlexDirectionRow || direction == FlexDirectionRowReverse
}

func FlexDirectionIsColumn(direction FlexDirection) bool {
	return direction == FlexDirectionColumn || direction == FlexDirectionColumnReverse
}

func LeadingMargin(node *Node, axis FlexDirection, widthSize float64) (float64, error) {
	if FlexDirectionIsRow(axis) && node.style.margin[EdgeStart].unit != UnitUndefined {
		return ValueResolve(&node.style.margin[EdgeStart], widthSize), nil
	}
	val, err := ComputedEdgeValue(node.style.margin, leading[axis], &Value{value: 0, unit: UnitUndefined})
	if err != nil {
		return 0, err
	}
	return ValueResolve(val, widthSize), nil
}
