import * as gospel from "./lib/index.js";

declare global {
    interface Window {
        gospel:any;
    }
}

window.gospel = gospel;

gospel.init();