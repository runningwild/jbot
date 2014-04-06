varying vec3 pos;
uniform float size;
uniform float edge;
uniform float coverage;

float alpha1d(float v, float ref) {
  if (v > 0.5) {
    v = 1.0 - v;
  }
  if (v > ref) {
    return 0.0;
  }
  if (v < ref - edge) {
    return 1.0;
  }
  return 1.0 - smoothstep(ref - edge, ref, v);
}

void main(void) {
  float xa1 = alpha1d(pos.x, size);
  float ya1 = alpha1d(pos.y, size);
  float xa2 = alpha1d(pos.x, coverage / 2.0);
  float ya2 = alpha1d(pos.y, coverage / 2.0);
  vec4 c = gl_Color;
  if (xa1 > 0.0) {
    xa1 = xa1 * ya2;
  }
  if (ya1 > 0.0) {
    ya1 = ya1 * xa2;
  }
  c.a = c.a * max(xa1, ya1);
  gl_FragColor = c;
  return;
}
