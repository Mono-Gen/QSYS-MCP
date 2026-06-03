# Q-SYS UCI CSS Styling Reference

This document is a cheat sheet for AI assistants and developers to refer to when customizing the Q-SYS User Control Interface (UCI) using CSS.

---

## 1. Q-SYS Specific CSS Selectors (Control Types)

The main selectors available in Q-SYS UCI are as follows:

| Selector | Target Control / Element |
| :--- | :--- |
| `button` | Various buttons (Mute, Trigger, Toggle, etc.) |
| `fader` | Faders for volume and other parameters |
| `knob` | Rotary knobs |
| `meter` | Level meters |
| `led` | LED indicators |
| `textbox` | Text boxes, text display elements |
| `textblock` | Static text blocks |
| `groupbox` | Border/background frames for grouping |
| `header` | UCI headers |
| `flexbox` | Container elements for layout positioning |
| `page` | UCI background / entire page |

---

## 2. Q-SYS Custom CSS Properties

These are special Q-SYS UCI stylesheet properties prefixed with `-qsc-`.

### Icon Font Properties
Draws built-in icon fonts (Foundation Icons or Material Icons) onto controls.

*   **`-qsc-icon-font`**: `foundation` (default) \| `material` \| custom font family name
*   **`-qsc-icon`**: Name of the icon to draw (e.g., `blind`, `brightness_1`) or Unicode character (e.g., `"\f167"`)
*   **`-qsc-icon-color`**: Icon color (e.g., `red`, `#FF0000`, `rgb(0,255,0)`)
*   **`-qsc-icon-align`**: Icon alignment within the control (`left` \| `right`)

### Render Styles (Meters, Knobs, Faders)
Customizes the drawing style of controls.

*   **`-qsc-render-style`**: `classic` (default) \| `filmstrip` (image frame animation) \| `layer` (multi-layered graphics)
*   **`-qsc-meter-indicator-class`**: Class name to use for drawing the level meter indicator
*   **`-qsc-meter-indicator-width`**: Width of the indicator (supports relative viewport units like `vh`, `vw`, etc.)
*   **`-qsc-meter-indicator-height`**: Height of the indicator (supports relative viewport units)

---

## 3. Supported Standard CSS Properties

The main standard CSS properties supported by Q-SYS UCI elements.
Note: Some properties may not be supported by specific elements.

*   **Background**: `background`, `background-color`, `background-image`, `background-position`, `background-size`
    *   Linear gradients are supported: `linear-gradient()`
*   **Borders**: `border`, `border-width`, `border-color`, `border-style`, `border-radius`
*   **Font/Text**: `color`, `font-family` (Note: Fallbacks are not supported, define a single font only), `font-size`, `font-style`, `font-weight`, `text-align`, `text-decoration`
*   **Padding/Margin/Size**: `padding`, `margin`, `width`, `height` (only effective inside a Flexbox)
*   **Opacity**: `opacity`

---

## 4. Pseudo-classes and Pseudo-states

Selectors used to apply styles based on control state.

*   **`:active` / `:value(1)`**: Applied when a button is ON or the control is active.
*   **`:pressed`**: Applied when the button or object is physically pressed (during click/touch).
*   **`:disabled`**: Applied when the control is disabled.
*   **`:value(...)`**: Used to apply styles based on the current `.Position` (0.0 to 1.0) of the control (primarily for filmstrip renders).

---

## 5. Coding Samples

### Sample A: MTR (Microsoft Teams Rooms) Style Buttons
```css
/* Primary Button (Light purple when ON, dark purple when OFF) */
.buttonprimary {
  background-color: #3b0066; /* OFF background */
  border-radius: 4px;
  color: #ffffff;
}

.buttonprimary:active, .buttonprimary:value(1) {
  background-color: #8a2be2; /* ON background */
}

/* Volume Mute Button with Icon */
.mute-button {
  -qsc-icon-font: material;
  -qsc-icon: "volume_off";
  -qsc-icon-color: #ff3b30;
  -qsc-icon-align: left;
}
```

### Sample B: Filmstrip Knob Definition
```css
/* Custom knob styling using a strip of frames */
.custom-knob {
  -qsc-render-style: filmstrip;
  background-image: url("images/knob_strip.png");
  background-size: 100% 5000%; /* If the image has 50 frames */
}
```

### Sample C: Custom Font Face
```css
@font-face {
  font-family: "My Custom Font";
  src: url(fonts/customfont.ttf);
}

textbox {
  font-family: "My Custom Font";
  font-size: 18px;
}
```
