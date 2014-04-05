varying vec3 pos;
uniform float size;
uniform float edge;
uniform float coverage;

float alpha1d(float v) {
  if (v > 0.5) {
    v = 1.0 - v;
  }
  if (v > size) {
    return 0.0;
  }
  if (v < size - edge) {
    return 1.0;
  }
  return 1.0 - smoothstep(size - edge, size, v);
}

void main(void) {
  float xa = alpha1d(pos.x);
  float ya = alpha1d(pos.y);
  vec4 c = gl_Color;
  c.a = c.a * max(xa, ya);
  gl_FragColor = c;
  return;
}
