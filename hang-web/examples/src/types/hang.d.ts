// Type augmentation so Solid's JSX knows about <hang-room>
declare module "solid-js" {
  namespace JSX {
    interface IntrinsicElements {
      "hang-room": any;
    }
  }
}

export {};
