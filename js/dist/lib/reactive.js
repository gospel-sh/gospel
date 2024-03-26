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
        this.current = this;
    }
    replace(value) {
        this.current = value;
        this.notify();
    }
    subscribe(handler) {
        this.subscribers.push(handler);
        return () => this.unsubscribe(handler);
    }
    unsubscribe(handler) {
        this.subscribers = this.subscribers.filter((a) => a !== handler);
    }
    notify() {
        for (const handler of this.subscribers) {
            handler(this.current);
        }
    }
}
export class Mapping extends Data {
    constructor(data) {
        super();
        const map = new Map([]);
        for (const [k, v] of Object.entries(data)) {
            if (v instanceof Data) {
                map.set(k, v);
                continue;
            }
            const value = new Literal(v);
            Object.defineProperty(this, k, {
                configurable: false,
                get: () => value,
            });
            map.set(k, value);
        }
        this.data = map;
    }
    _get(key) {
        const value = this.current.data.get(key);
        // if the value is undefiend, we return an undefined literal
        if (value == undefined)
            return new Literal(undefined);
        return value;
    }
    set(key, value) {
        this.current.data.set(key, value);
        this.notify();
    }
    get(key) {
        if (key instanceof (Literal))
            return derive((a, b) => {
                return a.current._get(b.current.get());
            }, [this, key]);
        return derive((a) => a.current._get(key), [this]);
    }
}
export class Literal extends Data {
    constructor(value) {
        super();
        this.value = value;
    }
    get() {
        return this.current.value;
    }
    set(value) {
        this.current.value = value;
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
        return derive((a) => a.list[index], [this]);
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
    subscribe(value, handler) {
        this.unsubscribe = value.subscribe(handler);
    }
    unmount() {
        if (this.unsubscribe !== undefined) {
            this.unsubscribe();
        }
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
        let value = "unknown";
        if (this.value instanceof Data) {
            let cv = this.value;
            value = "data";
            if (cv instanceof Literal)
                value = cv.get();
        }
        else {
            value = this.value;
        }
        const node = document.createTextNode(value);
        if (this.value instanceof Literal)
            this.subscribe(this.value, (value) => {
                node.data = value.get();
            });
        super.mount(parent, node);
    }
}
export class Tag extends Component {
    constructor(tag, ...children) {
        super();
        this.tag = tag;
        this.children = children;
    }
    unmount() {
        for (const child of this.children) {
            if (child instanceof Component) {
                child.unmount();
            }
        }
        super.unmount();
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
        const handler = (condition) => {
            const newValue = condition.get();
            if (!newValue && value && this.view !== undefined) {
                value = false;
                this.view.unmount();
                if (this.viewFunction !== undefined)
                    this.view = undefined;
            }
            else if (newValue && !value) {
                if (this.view === undefined) {
                    // here we need to mark all subscribers as conditional
                    // on a given variable, which will then give us the
                    // ability to discard all updates for these variables
                    this.view = this.viewFunction();
                }
                value = true;
                this.view.render(parent);
            }
        };
        this.condition.subscribe(handler);
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
// Derivation: Calls a function that produces some data, which depends
// on some other data. If one of the dependencies changes, the data will
// be updated automatically.
export function derive(f, dependencies) {
    const data = f(...dependencies);
    for (const [index, dependency] of dependencies.entries()) {
        const unsubscribe = dependency.subscribe((value) => {
            dependencies[index] = value;
            try {
                const newData = f(...dependencies);
                data.replace(newData);
            }
            catch (e) {
                unsubscribe();
                console.error(e);
                return;
            }
        });
    }
    return data;
}
export function not(value) {
    return derive((a) => new Literal(!a.get()), [value]);
}
export function is(a, constructor) {
    return derive((b) => {
        return new Literal(b instanceof constructor);
    }, [a]);
}
//# sourceMappingURL=reactive.js.map