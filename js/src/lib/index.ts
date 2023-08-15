function handleClick(e: Event){
	let target = e.target;
	if (target instanceof HTMLElement && target.tagName !== 'a'){
		target = target.closest('a');
	}

	if (target === null){
		console.log("not a link");
		return;
	}

	const link = (target as HTMLAnchorElement).href;

	console.log("Link:", link);

	e.preventDefault();

	navigateTo(link, true);

}

function handlePopState(_: PopStateEvent){
	console.log(`Going back to ${document.location.href}`)
	navigateTo(document.location.href, false);
}

async function navigateTo(link: string, push: boolean){

	const response = await fetch(link);
	const doc = new DOMParser().parseFromString(await response.text(), "text/html");

	document.replaceChild(doc.all[0], document.all[0]);

	if (push){
		history.pushState(null, "", link);
	}

}

function addEventListeners(){

	addEventListener('click', handleClick);
	addEventListener('popstate', handlePopState);

}

function initGospel() {
	console.log("initializing gospel")
	const submittables = document.querySelectorAll("[gospel-onSubmit]");
	for(const [_, submittable] of submittables.entries()){
		console.log("adding onSubmit handler...");
		(submittable as HTMLFormElement).onsubmit = (_: Event) => {
			console.log("submitting...");
			// e.preventDefault();
		}
	}

	addEventListeners();

}

export function init() {

	document.addEventListener('DOMContentLoaded', initGospel, false);
}

