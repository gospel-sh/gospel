
function initGospel() {
	console.log("initializing gospel")
	const submittables = document.querySelectorAll("[gospel-onSubmit]");
	for(const [_, submittable] of submittables.entries()){
		console.log("adding onSubmit handler...");
		(submittable as HTMLFormElement).onsubmit = (e: Event) => {
			console.log("submitting...");
			e.preventDefault();
		}
	}
}

export function init() {

	document.addEventListener('DOMContentLoaded', initGospel, false);

}

