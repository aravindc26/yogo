package yoga

type Direction int

const (
	DirectionInherit Direction = iota
	DirectionLTR
	DirectionRTL
)

type MeasureMode int

const (
	MeasureModeUndefined MeasureMode = iota
	MeasureModeExactly
	MeasureModeAtmost
)

type FlexDirection int

const (
	FlexDirectionColumn FlexDirection = iota
	FlexDirectionColumnReverse
	FlexDirectionRow
	FlexDirectionRowReverse
)

func (f FlexDirection) String() string {
	switch f {
	case FlexDirectionColumn:
		return "column"
	case FlexDirectionColumnReverse:
		return "column-reverse"
	case FlexDirectionRow:
		return "row"
	case FlexDirectionRowReverse:
		return "row-reverse"
	}
	return ""
}

type Justify int

const (
	JustifyFlexStart Justify = iota
	JustifyCenter
	JustifyFlexEnd
	JustifySpaceBetween
	JustifySpaceAround
)

func (j Justify) String() string {
	switch j {
	case JustifyCenter:
		return "center"
	case JustifyFlexEnd:
		return "flex-end"
	case JustifySpaceBetween:
		return "space-between"
	case JustifySpaceAround:
		return "space-around"
	case JustifyFlexStart:
		return "flex-start"
	}
	return ""
}

type Align int

const (
	AlignAuto Align = iota
	AlignFlexStart
	AlignCenter
	AlignFlexEnd
	AlignStretch
	AlignBaseLine
)

func (a Align) String() string {
	switch a {
	case AlignCenter:
		return "center"
	case AlignFlexEnd:
		return "flex-end"
	case AlignStretch:
		return "stretch"
	case AlignBaseLine:
		return "base-line"
	case AlignFlexStart:
		return "flex-start"
	case AlignAuto:
		return "auto"
	}
	return ""
}

type PositionType int

const (
	PositionTypeRelative PositionType = iota
	PositionTypeAbsolute
)

func (p PositionType) String() string {
	switch p {
	case PositionTypeRelative:
		return "relative"
	case PositionTypeAbsolute:
		return "absolute"
	}
	return ""
}

type Wrap int

const (
	WrapNoWrap Wrap = iota
	WrapWrap
)

type Overflow int

const (
	OverflowVisible Overflow = iota
	OverflowHidden
	OverflowScroll
)

func (o Overflow) String() string {
	switch o {
	case OverflowVisible:
		return "visible"
	case OverflowHidden:
		return "hidden"
	case OverflowScroll:
		return "scroll"
	}
	return ""
}

type Unit int

const (
	UnitUndefined Unit = iota
	UnitPixel
	UnitPercent
)

func (u Unit) String() string {
	switch u {
	case UnitPixel:
		return "px"
	case UnitPercent:
		return "%"
	}
	return ""
}

type Edge int

const (
	EdgeLeft Edge = iota
	EdgeTop
	EdgeRight
	EdgeBottom
	EdgeStart
	EdgeEnd
	EdgeHorizontal
	EdgeVertical
	EdgeAll
)

type Dimension int

const (
	DimensionWidth Dimension = iota
	DimensionHeight
)

type PrintOptions int

const (
	PrintOptionsLayout PrintOptions = 1 << iota
	PrintOptionsStyle
	PrintOptionsChildren
)
