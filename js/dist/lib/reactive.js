/*
enum Change {
    Insert = 1,
    Remove,
    Swap,
    Update,
}
*/
export class Data {
    constructor() {
        this.subscribers = [];
    }
    replace(_) {
    }
    subscribe(handler) {
        this.subscribers.push(handler);
    }
    unsubscribe(handler) {
        this.subscribers = this.subscribers.filter((a) => a !== handler);
    }
    notify() {
        for (const handler of this.subscribers) {
            handler();
        }
    }
}
export class Mapping extends Data {
    constructor(data) {
        super();
        const map = new Map([]);
        for (const [k, v] of Object.entries(data)) {
            const value = new Literal(v);
            Object.defineProperty(this, k, {
                configurable: false,
                get: () => value,
            });
            map.set(k, value);
        }
        this.data = map;
    }
    get(key) {
        return this.data.get(key);
    }
}
export class Literal extends Data {
    constructor(value) {
        super();
        this.value = value;
    }
    replace(data) {
        if (data instanceof Literal)
            this.set(data.get());
    }
    get() {
        return this.value;
    }
    set(value) {
        this.value = value;
        this.notify();
    }
}
export class List extends Data {
    constructor(data) {
        super();
        const list = [];
        for (const item of data) {
            const value = new Literal(item);
            list.push(value);
        }
        this.list = list;
    }
    get(index) {
        console.log(this.list[index]);
        return this.list[index];
    }
    length() {
        return this.list.length;
    }
    *[Symbol.iterator]() {
        for (const item of this.list) {
            yield item;
        }
    }
}
class Component {
    mount(parent, node) {
        this.parent = parent;
        this.node = node;
        parent.appendChild(node);
    }
    unmount() {
        if (this.parent === undefined || this.node === undefined)
            return;
        this.parent.removeChild(this.node);
        this.parent = undefined;
        this.node = undefined;
    }
}
export class Attribute extends Component {
    constructor(name, value) {
        super();
        this.name = name;
        this.value = value;
    }
    render(parent) {
        const set = () => parent.setAttribute(this.name, this.value.get());
        set();
        this.value.subscribe(() => {
            set();
        });
    }
}
export class TextNode extends Component {
    constructor(value) {
        super();
        this.value = value;
    }
    render(parent) {
        let value;
        if (this.value instanceof Literal) {
            value = this.value.get();
        }
        else {
            value = this.value;
        }
        const node = document.createTextNode(value);
        if (this.value instanceof Literal)
            this.value.subscribe(() => node.data = this.value.get());
        super.mount(parent, node);
    }
}
export class Tag extends Component {
    constructor(tag, ...children) {
        super();
        this.tag = tag;
        this.children = children;
    }
    render(parent) {
        // we create the element
        const node = document.createElement(this.tag);
        // we render the children
        for (const child of this.children) {
            if (child instanceof Component) {
                // this is an attribute
                child.render(node);
            }
            else {
                console.error("unknown child element:", child);
            }
        }
        this.mount(parent, node);
    }
}
export class ListView extends Component {
    constructor(list, view) {
        super();
        this.list = list;
        this.view = view;
    }
    render(parent) {
        for (const item of this.list) {
            const tag = this.view(item);
            tag.render(parent);
        }
    }
}
export class IfView extends Component {
    constructor(condition, view) {
        super();
        if (view instanceof Component)
            this.view = view;
        else
            this.viewFunction = view;
        this.condition = condition;
    }
    render(parent) {
        let value = this.condition.get();
        this.condition.subscribe(() => {
            const newValue = this.condition.get();
            if (!newValue && value && this.view !== undefined) {
                value = false;
                this.view.unmount();
                if (this.viewFunction !== undefined)
                    this.view = undefined;
            }
            else if (newValue && !value) {
                if (this.view === undefined)
                    this.view = this.viewFunction();
                value = true;
                this.view.render(parent);
            }
        });
        if (value === true) {
            if (this.view === undefined)
                this.view = this.viewFunction();
            this.view.render(parent);
        }
    }
}
// Text helper
export const Text = (value) => {
    return new TextNode(value);
};
// Tags
function tag(tag) {
    return (...children) => new Tag(tag, ...children);
}
export const H1 = tag("h1");
export const Span = tag("span");
export const Div = tag("div");
export const P = tag("p");
export const Ul = tag("ul");
export const Li = tag("li");
// Iterators
export function For(list, view) {
    return new ListView(list, view);
}
export function If(condition, view) {
    return new IfView(condition, view);
}
// Derivation
export function derive(f, dependencies) {
    const data = f();
    for (const dependency of dependencies) {
        dependency.subscribe(() => {
            console.log("updating derived value...");
            const newData = f();
            data.replace(newData);
        });
    }
    return data;
}
//# sourceMappingURL=reactive.js.map