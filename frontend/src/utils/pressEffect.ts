/**
 * Button press feedback now lives in CSS (`_components.scss` :active scale).
 * Kept as a no-op so existing bootstrap call sites stay valid.
 */
export function installPressEffect(_root: Document = document): void {
  // CSS handles press feedback; GSAP is no longer loaded on first paint.
}
