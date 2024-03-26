
/*
enum Change {
	Insert = 1,
	Remove,
	Swap,
	Update,
}
*/

interface Renderable {
	render(parent: HTMLElement): void
	mount(parent: HTMLElement, node: Node): void
	unmount(parent: HTMLElement): void
}

interface Replaceable {
	replace(data: Data): void
}

type NotifyHandler = (notifier: Data) => void;

interface Observable {
	subscribe(handler: NotifyHandler): () => void
}

export class Data implements Observable, Replaceable {

	subscribers: NotifyHandler[]
	current: this

	constructor(){
		this.subscribers = []
		this.current = this
	}

	replace(value: this): void {
		this.current = value
		this.notify()
	}

	subscribe(handler: NotifyHandler): () => void {
		this.subscribers.push(handler)
		return () => this.unsubscribe(handler)
	}

	unsubscribe(handler: NotifyHandler): void {
		this.subscribers = this.subscribers.filter((a) => a !== handler)
	}

	notify() {
		for(const handler of this.subscribers){
			handler(this.current)
		}
	}
}

export class Mapping extends Data {

	data: Map<string, Data>

	constructor(data: Object){
		super()

		const map: Map<string, Data> = new Map([])

		for(const [k, v] of Object.entries(data)){

			if (v instanceof Data){
				map.set(k, v)
				continue
			}

			const value = new Literal(v)

			Object.defineProperty(this, k, {
				configurable: false,
				get: () => value,
			})

			map.set(k, value)
		}

		this.data = map

	}

	private _get(key: string): Data {
		const value = this.current.data.get(key)
		// if the value is undefiend, we return an undefined literal
		if (value == undefined)
			return new Literal(undefined)
		return value
	}

	set(key: string, value: Data) {
		this.current.data.set(key, value)
		this.notify()
	}

	get(key: string | Literal<string>): Data {
		if (key instanceof Literal<string>)
			return derive((a: this, b: Literal<string>) => {
				return a.current._get(b.current.get())
			}, [this, key])
		return derive((a: this) => a.current._get(key), [this])
	}
}

export class Literal<T=any> extends Data {

	value: T

	constructor(value: T){
		super()
		this.value = value
	}

	get(): T {
		return this.current.value
	}

	set(value: T): void {
		this.current.value = value
		this.notify()
	}
}



export class List extends Data {

	list: Data[]

	constructor(data: any[]){
		super()

		const list: Data[] = []

		for(const item of data){
			const value = new Literal(item)
			list.push(value)
		}

		this.list = list
	}


	get(index: number): Data {
		return derive((a: this) => a.list[index], [this])
	}

	length(): number {
		return this.list.length
	}

	*[Symbol.iterator](): IterableIterator<Data> {
		for(const item of this.list){
			yield item
		}
	}
}

abstract class Component implements Renderable {

	unsubscribe: () => void | undefined
	parent?: HTMLElement
	node?: Node

	abstract render(parent: HTMLElement): void

	mount(parent: HTMLElement, node: Node){
		this.parent = parent
		this.node = node
		parent.appendChild(node)

	}

	subscribe(value: Data, handler: NotifyHandler) {
		this.unsubscribe = value.subscribe(handler)
	}

	unmount(){

		if (this.unsubscribe !== undefined){
			this.unsubscribe()
		}

		if (this.parent === undefined || this.node === undefined)
			return

		this.parent.removeChild(this.node)
		this.parent = undefined
		this.node = undefined
	}

}

export class Attribute extends Component {
	name: string
	value: Literal

	constructor(name: string, value: Literal){
		super()
		this.name = name
		this.value = value
	}

	render(parent: HTMLElement){

		const set = () => parent.setAttribute(this.name, this.value.get())

		set()
		
		this.value.subscribe(() => {
			set()
		})
	}
}

export class TextNode extends Component {

	value: Data | string

	constructor(value: Data | string) {
		super()
		this.value = value
	}

	render(parent: HTMLElement) {

		let value = "unknown"

		if (this.value instanceof Data){
			let cv = this.value
			value = "data"
			if (cv instanceof Literal)
				value = cv.get()
		} else {
			value = this.value
		}

		const node = document.createTextNode(value)

		if (this.value instanceof Literal)
			this.subscribe(this.value, (value: Literal) => {
				node.data = value.get()
			})

		super.mount(parent, node)
	}
}

export class Tag extends Component {

	tag: string
	children: any[]

	constructor(tag: string, ...children: any[]){
		super()
		this.tag = tag
		this.children = children
	}

	unmount(){
		for(const child of this.children){
			if (child instanceof Component){
				child.unmount()
			}
		}
		super.unmount()
	}

	render(parent: HTMLElement) {
		// we create the element
		const node = document.createElement(this.tag)

		// we render the children
		for(const child of this.children){

			if (child instanceof Component) {
				// this is an attribute
				child.render(node);
			} else {				
				console.error("unknown child element:", child)
			}
		}

		this.mount(parent, node)
	}
}

export class ListView extends Component {

	list: List
	view: (item: Data) => Tag

	constructor(list: List, view: (item: Data) => Tag){
		super()
		this.list = list
		this.view = view
	}

	render(parent: HTMLElement){
		for(const item of this.list){

			const tag = this.view(item)
			tag.render(parent)

		}
	}
}

export class IfView extends Component {

	condition: Literal
	viewFunction?: () => Component
	view?: Component

	constructor(condition: Literal, view: Component | (() => Component)){
		super()

		if (view instanceof Component)
			this.view = view
		else
			this.viewFunction = view

		this.condition = condition
	}

	render(parent: HTMLElement){

		let value = this.condition.get()

		const handler = (condition: Literal) => {

			const newValue = condition.get()

			if (!newValue && value && this.view !== undefined){

				value = false
				this.view.unmount()
				
				if (this.viewFunction !== undefined)
					this.view = undefined

			} else if (newValue && !value){

				if (this.view === undefined){
					// here we need to mark all subscribers as conditional
					// on a given variable, which will then give us the
					// ability to discard all updates for these variables
					this.view = this.viewFunction!()
				}

				value = true
				this.view.render(parent)
			}
		}

		this.condition.subscribe(handler)

		if(value === true){

			if (this.view === undefined)
				this.view = this.viewFunction!()

			this.view.render(parent)
		}
	}
}

// Text helper

export const Text = (value: Literal | string): TextNode => {
	return new TextNode(value)
}

// Tags

function tag(tag: string): (...children: any[]) => Tag {
	return (...children: any[]) => new Tag(tag, ...children)
}

export const H1 = tag("h1");
export const Span = tag("span");
export const Div = tag("div");
export const P = tag("p");
export const Ul = tag("ul");
export const Li = tag("li");

// Iterators

export function For(list: List, view: (item: Data) => Tag): Component {
	return new ListView(list, view)
}

export function If(condition: Literal, view: Component): Component {
	return new IfView(condition, view)
}

// Derivation: Calls a function that produces some data, which depends
// on some other data. If one of the dependencies changes, the data will
// be updated automatically.
export function derive<T extends Data, G extends Data[]>(f: (...args: G) => T, dependencies: [...G]): T {
	const data = f(...dependencies)

	for(const [index, dependency] of dependencies.entries()){
		const unsubscribe = dependency.subscribe((value: Data) => {
			dependencies[index] = value
			try {
				const newData = f(...dependencies)
				data.replace(newData)
			} catch(e) {
				unsubscribe()
				console.error(e)
				return
			}
		})
	}

	return data
}

export function not(value: Literal<boolean>): Literal<boolean> {
	return derive((a: typeof value) => new Literal<boolean>(!a.get()), [value])
}

export function is<T>(a: any, constructor: { new (...args: any[]): T }): Literal<boolean> {
    return derive((b: typeof a) => {
    	return new Literal<boolean>(b instanceof constructor)
    }, [a]);
}
