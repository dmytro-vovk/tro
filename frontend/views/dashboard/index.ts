import App from '../../system';
import {$html, $onClick, $text} from '../index';

type response = {
	message: string
}

type stream = {
	value: string
}

export default class Dashboard {
	constructor(private readonly app: App) {
	}

	public render(selector: string): void {
		$html(selector, `
			<div class="row">
				<div class="col-4">
					<div class="card card-primary card-outline card-outline-tabs">
						<div class="card-header">
							<h3 class="card-title">Example request/response</h3>
						</div>
						<div class="card-body" id="response"></div>
						<div class="card-footer">
							<button id="ping" type="button" class="btn btn-success float-right">Click me</button>
						</div>
					</div>
				</div>
				<div class="col-4">
					<div class="card card-primary card-outline card-outline-tabs">
						<div class="card-header">
							<h3 class="card-title">Server stream</h3>
						</div>
						<div class="card-body" id="stream"></div>
						<div class="card-footer"></div>
					</div>
				</div>
			</div>`);
		this.setupExampleHandler();
		this.subscribeToServerStream();
	}

	private subscribeToServerStream() {
		this.app.subscribe("example.stream", (data) => {
			const s = data as unknown as stream;
			$text("#stream", s.value);
		})
	}

	private setupExampleHandler() {
		$onClick(
			"#ping",
			() => {
				this.app.call(
						"example.method",
						{
							message: "hello"
						}
					)
					.then(
						(data) => {
							const r = data as unknown as response;
							$text("#response", r.message);
						},
						(error) => this.app.error(error),
					)
			}
		);
	}
}
