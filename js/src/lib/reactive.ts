
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

type NotifyHandler = () => void;

interface Observable {
	subscribe(handler: NotifyHandler): void
	unsubscribe(handler: NotifyHandler): void
}

export class Data implements Observable, Replaceable {

	subscribers: NotifyHandler[]

	constructor(){
		this.subscribers = []
	}

	replace(_: Data): void {

	}

	subscribe(handler: NotifyHandler): void {
		this.subscribers.push(handler)
	}

	unsubscribe(handler: NotifyHandler): void {
		this.subscribers = this.subscribers.filter((a) => a !== handler)
	}

	notify() {
		for(const handler of this.subscribers){
			handler()
		}
	}


}

export class Mapping extends Data {

	data: Map<string, Data>

	constructor(data: Object){
		super()

		const map: Map<string, Data> = new Map([])

		for(const [k, v] of Object.entries(data)){

			const value = new Literal(v)

			Object.defineProperty(this, k, {
				configurable: false,
				get: () => value,
			})

			map.set(k, value)
		}

		this.data = map

	}

	get(key: string): Data | undefined {
		return this.data.get(key)
	}
}

export class Literal extends Data {

	value: any

	constructor(value: any){
		super()
		this.value = value
	}

	replace(data: Data): void {
		if (data instanceof Literal)
			this.set(data.get())
	}

	get(): any {
		return this.value
	}

	set(value: any): void {
		this.value = value
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
		console.log(this.list[index])
		return this.list[index]
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

	parent?: HTMLElement
	node?: Node

	abstract render(parent: HTMLElement): void

	mount(parent: HTMLElement, node: Node){
		this.parent = parent
		this.node = node

		parent.appendChild(node)

	}

	unmount(){
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

	value: Literal | string

	constructor(value: Literal | string) {
		super()
		this.value = value
	}

	render(parent: HTMLElement) {

		let value

		if (this.value instanceof Literal){
			value = this.value.get()
		} else {
			value = this.value
		}

		const node = document.createTextNode(value)

		if (this.value instanceof Literal)
			this.value.subscribe(() => node.data = (this.value as Literal).get())

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

		this.condition.subscribe(() => {

			const newValue = this.condition.get()


			if (!newValue && value && this.view !== undefined){
	
				value = false
				this.view.unmount()

				if (this.viewFunction !== undefined)
					this.view = undefined

			} else if (newValue && !value){

				if (this.view === undefined)
					this.view = this.viewFunction!()

				value = true
				this.view.render(parent)
			}
		})

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

// Derivation

export function derive(f: () => Data, dependencies: Data[]): Data {
	const data = f()

	for(const dependency of dependencies){
		dependency.subscribe(() => {
			console.log("updating derived value...")
			const newData = f()
			data.replace(newData)
		})
	}

	return data
}