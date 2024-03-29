package huego

import (
	"context"
	"image/color"
	"math"
)

// Light represents a bridge light https://developers.meethue.com/documentation/lights-api
type Light struct {
	State            *State `json:"state,omitempty"`
	Type             string `json:"type,omitempty"`
	Name             string `json:"name,omitempty"`
	ModelID          string `json:"modelid,omitempty"`
	ManufacturerName string `json:"manufacturername,omitempty"`
	UniqueID         string `json:"uniqueid,omitempty"`
	SwVersion        string `json:"swversion,omitempty"`
	SwConfigID       string `json:"swconfigid,omitempty"`
	ProductName      string `json:"productname,omitempty"`
	ID               int    `json:"-"`
	bridge           *Bridge
}

// State defines the attributes and properties of a light
type State struct {
	On             bool      `json:"on"`
	Bri            uint8     `json:"bri,omitempty"`
	Hue            uint16    `json:"hue,omitempty"`
	Sat            uint8     `json:"sat,omitempty"`
	Xy             []float32 `json:"xy,omitempty"`
	Ct             uint16    `json:"ct,omitempty"`
	Alert          string    `json:"alert,omitempty"`
	Effect         string    `json:"effect,omitempty"`
	TransitionTime uint16    `json:"transitiontime,omitempty"`
	BriInc         int       `json:"bri_inc,omitempty"`
	SatInc         int       `json:"sat_inc,omitempty"`
	HueInc         int       `json:"hue_inc,omitempty"`
	CtInc          int       `json:"ct_inc,omitempty"`
	XyInc          int       `json:"xy_inc,omitempty"`
	ColorMode      string    `json:"colormode,omitempty"`
	Reachable      bool      `json:"reachable,omitempty"`
	Scene          string    `json:"scene,omitempty"`
}

// NewLight defines a list of lights discovered the last time the bridge performed a light discovery.
// Also stores the timestamp the last time a discovery was performed.
type NewLight struct {
	Lights   []string
	LastScan string `json:"lastscan"`
}

// SetState sets the state of the light to s.
func (l *Light) SetState(s State) error {
	return l.SetStateContext(context.Background(), s)
}

// SetStateContext sets the state of the light to s.
func (l *Light) SetStateContext(ctx context.Context, s State) error {
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, s)
	if err != nil {
		return err
	}
	l.State = &s
	return nil
}

// Off sets the On state of one light to false, turning it off
func (l *Light) Off() error {
	return l.OffContext(context.Background())
}

// OffContext sets the On state of one light to false, turning it off
func (l *Light) OffContext(ctx context.Context) error {
	state := State{On: false}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, state)
	if err != nil {
		return err
	}
	l.State.On = false
	return nil
}

// On sets the On state of one light to true, turning it on
func (l *Light) On() error {
	return l.OnContext(context.Background())
}

// OnContext sets the On state of one light to true, turning it on
func (l *Light) OnContext(ctx context.Context) error {
	state := State{On: true}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, state)
	if err != nil {
		return err
	}
	l.State.On = true
	return nil
}

// IsOn returns true if light state On property is true
func (l *Light) IsOn() bool {
	return l.State.On
}

// Rename sets the name property of the light
func (l *Light) Rename(new string) error {
	return l.RenameContext(context.Background(), new)
}

// RenameContext sets the name property of the light
func (l *Light) RenameContext(ctx context.Context, new string) error {
	update := Light{Name: new}
	_, err := l.bridge.UpdateLightContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.Name = new
	return nil
}

// Bri sets the light brightness state property
func (l *Light) Bri(new uint8) error {
	return l.BriContext(context.Background(), new)
}

// BriContext sets the light brightness state property
func (l *Light) BriContext(ctx context.Context, new uint8) error {
	update := State{On: true, Bri: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Bri = new
	l.State.On = true
	return nil
}

// Hue sets the light hue state property (0-65535)
func (l *Light) Hue(new uint16) error {
	return l.HueContext(context.Background(), new)
}

// HueContext sets the light hue state property (0-65535)
func (l *Light) HueContext(ctx context.Context, new uint16) error {
	update := State{On: true, Hue: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Hue = new
	l.State.On = true
	return nil
}

// Sat sets the light saturation state property (0-254)
func (l *Light) Sat(new uint8) error {
	return l.SatContext(context.Background(), new)
}

// SatContext sets the light saturation state property (0-254)
func (l *Light) SatContext(ctx context.Context, new uint8) error {
	update := State{On: true, Sat: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Sat = new
	l.State.On = true
	return nil
}

// Xy sets the x and y coordinates of a color in CIE color space. (0-1 per value)
func (l *Light) Xy(new []float32) error {
	return l.XyContext(context.Background(), new)
}

// XyContext sets the x and y coordinates of a color in CIE color space. (0-1 per value)
func (l *Light) XyContext(ctx context.Context, new []float32) error {
	update := State{On: true, Xy: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Xy = new
	l.State.On = true
	return nil
}

// Ct sets the light color temperature state property
func (l *Light) Ct(new uint16) error {
	return l.CtContext(context.Background(), new)
}

// CtContext sets the light color temperature state property
func (l *Light) CtContext(ctx context.Context, new uint16) error {
	update := State{On: true, Ct: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Ct = new
	l.State.On = true
	return nil
}

// Col sets the light color as RGB (will be converted to xy)
func (l *Light) Col(new color.Color) error {
	return l.ColContext(context.Background(), new)
}

// ColContext sets the light color as RGB (will be converted to xy)
func (l *Light) ColContext(ctx context.Context, new color.Color) error {
	xy, bri := ConvertRGBToXy(new)

	update := State{On: true, Xy: xy, Bri: bri}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Xy = xy
	l.State.Bri = bri
	l.State.On = true
	return nil
}

// TransitionTime sets the duration of the transition from the light’s current state to the new state
func (l *Light) TransitionTime(new uint16) error {
	return l.TransitionTimeContext(context.Background(), new)
}

// TransitionTimeContext sets the duration of the transition from the light’s current state to the new state
func (l *Light) TransitionTimeContext(ctx context.Context, new uint16) error {
	update := State{On: l.State.On, TransitionTime: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.TransitionTime = new
	return nil
}

// Effect the dynamic effect of the light, currently “none” and “colorloop” are supported
func (l *Light) Effect(new string) error {
	return l.EffectContext(context.Background(), new)
}

// EffectContext the dynamic effect of the light, currently “none” and “colorloop” are supported
func (l *Light) EffectContext(ctx context.Context, new string) error {
	update := State{On: true, Effect: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Effect = new
	l.State.On = true
	return nil
}

// Alert makes the light blink in its current color. Supported values are:
// “none” – The light is not performing an alert effect.
// “select” – The light is performing one breathe cycle.
// “lselect” – The light is performing breathe cycles for 15 seconds or until alert is set to "none".
func (l *Light) Alert(new string) error {
	return l.AlertContext(context.Background(), new)
}

// AlertContext makes the light blink in its current color. Supported values are:
// “none” – The light is not performing an alert effect.
// “select” – The light is performing one breathe cycle.
// “lselect” – The light is performing breathe cycles for 15 seconds or until alert is set to "none".
func (l *Light) AlertContext(ctx context.Context, new string) error {
	update := State{On: true, Alert: new}
	_, err := l.bridge.SetLightStateContext(ctx, l.ID, update)
	if err != nil {
		return err
	}
	l.State.Effect = new
	l.State.On = true
	return nil
}

// ConvertRGBToXy converts a given RGB color to the xy color of the ligth.
// implemented as in https://developers.meethue.com/develop/application-design-guidance/color-conversion-formulas-rgb-to-xy-and-back/
func ConvertRGBToXy(newcolor color.Color) ([]float32, uint8) {
	r, g, b, _ := newcolor.RGBA()
	rf := float64(r) / 65536.0
	gf := float64(g) / 65536.0
	bf := float64(b) / 65536.0

	rf = gammaCorrect(rf)
	gf = gammaCorrect(gf)
	bf = gammaCorrect(bf)

	X := float32(rf*0.649926 + gf*0.103455 + bf*0.197109)
	Y := float32(rf*0.234327 + gf*0.743075 + bf*0.022598)
	Z := float32(rf*0.0000000 + gf*0.053077 + bf*1.035763)

	x := X / (X + Y + Z)
	y := Y / (X + Y + Z)

	xy := []float32{x, y}
	return xy, uint8(Y * 254)
}

func gammaCorrect(value float64) float64 {
	if value > 0.04045 {
		return math.Pow((value+0.055)/(1.0+0.055), 2.4)
	}
	return (value / 12.92)
}
