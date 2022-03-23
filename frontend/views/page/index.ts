import {$html} from '../index';

export interface Renderable {
	render(selector: string): void;
}

export default class Page {
	constructor(
		private readonly title: string,
		private readonly content: Renderable,
	) {}

	public render(selector: string): void {
		$html(selector, `
            <section class="content-header">
                <div class="container-fluid">
                    <div class="row mb-2">
                        <div class="col-sm-12">
                            <h1>${this.title}</h1>
                        </div>
                    </div>
                </div>
            </section>
            <section class="content mb-5">
                <div class="container-fluid" id="content"></div>
            </section>`);
		this.content.render('#content');
	}
}
