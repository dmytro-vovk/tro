export function $$(selector: string): HTMLElement {
	return document.querySelector(selector);
}

export function $html(element: string | HTMLElement, content: string | string[], glue = ''): void {
	if (typeof element === 'string') {
		element = $$(element);
	}
	if (Array.isArray(content)) {
		content = content.join(glue);
	}
	element.innerHTML = content;
}

export function $text(element: string | HTMLElement, content: number | string): void {
	if (typeof element === 'string') {
		element = $$(element);
	}
	if (typeof content === 'number') {
		content = content.toString();
	}
	element.innerText = content;
}

export function $onClick<K extends keyof HTMLElementEventMap>(element: HTMLElement | string, callback: (this: HTMLSelectElement, ev: HTMLElementEventMap[K]) => any): void {
	if (typeof element === 'string') {
		element = document.querySelector(element) as HTMLElement;
	}
	element.addEventListener('click', callback);
}

export function $onChange<K extends keyof HTMLElementEventMap>(element: string | HTMLElement, callback: (this: HTMLSelectElement, ev: HTMLElementEventMap[K]) => any): void {
	if (typeof element === 'string') {
		element = $$(element) as HTMLElement;
	}
	element.addEventListener('change', callback);
}

export function $newElement(tagName: string, options = {}): HTMLElement {
	const e = document.createElement(tagName);
	for (const prop of Object.keys(options)) {
		e[prop] = options[prop];
	}
	return e
}
