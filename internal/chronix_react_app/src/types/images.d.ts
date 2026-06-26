// src/types/images.d.ts
declare module "*.png";
declare module "*.svg" {
    const src: string;
    export default src;
}
declare module "*.svg?react" {
    import type * as React from "react";
    const ReactComponent: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
    export default ReactComponent;
}