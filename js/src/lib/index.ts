function handleClick(e: Event){
	let target = e.target;
	if (target instanceof HTMLElement && target.tagName !== 'a'){
		target = target.closest('a');
	}

	if (target === null){
		return;
	}

	const link = (target as HTMLAnchorElement).href;
	e.preventDefault();

	navigateTo(link, true);

}

function handlePopState(_: PopStateEvent){
	navigateTo(document.location.href, false);
}

async function handleOnSubmit(e: SubmitEvent) {
	e.preventDefault();

	const form = e.target as HTMLFormElement

	// we create the form data
	const formData = new FormData(form);

	if (e.submitter){

		const button = e.submitter as HTMLButtonElement

		if (button.name !== "")
			formData.append(button.name, button.value);
	}

	// we need this instead of using form.action as that can be overwritten
	// if a field named 'action' is present in the form...
	const action = Object.getOwnPropertyDescriptor(HTMLFormElement.prototype, 'action')!.get!.call(form)

	const response = await fetch(action, {
		body: formData,
		method: form.method,
	})

	replaceDom(response.url, await response.text(), response.redirected);

}

async function navigateTo(link: string, push: boolean){

	const response = await fetch(link);

	replaceDom(response.url, await response.text(), push || response.redirected);
}

function replaceDom(link: string, text: string, push: boolean) {

	const doc = new DOMParser().parseFromString(text, "text/html");

	// we capture the scroll position
	const scrollX = window.scrollX;
	const scrollY = window.scrollY;	

	document.replaceChild(doc.all[0], document.all[0]);

	// we restore the scroll position
	window.scroll(scrollX, scrollY);

	if (push){
		history.pushState(null, "", link);
	}

	// we add the event handlers...
	initDocument();

}

function addEventListeners(){

	addEventListener('click', handleClick);
	addEventListener('popstate', handlePopState);

}

function initDocument(){
	const submittables = document.querySelectorAll("[gospel-onSubmit]");

	for(const [_, submittable] of submittables.entries()){
		console.log("adding onSubmit handler...");
		(submittable as HTMLFormElement).onsubmit = handleOnSubmit;
	}
}


function initGospel() {
	initDocument();
	addEventListeners();
}

export function init() {

	document.addEventListener('DOMContentLoaded', initGospel, false);
}

