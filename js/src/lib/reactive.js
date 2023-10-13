export class Data {
}
export class Map extends Data {
}
export class Literal extends Data {
    get() {
        return "foo";
    }
}
export class Array extends Data {
}
export class Component {
}
export class H1 extends Component {
    constructor(...data) {
        this.children = data;
    }
    render(parent) {
        const h1 = document.createElement("h1");
        h1.innerText = this.data[0].get();
        parent.appendChild(h1);
    }
}
//# sourceMappingURL=reactive.js.map