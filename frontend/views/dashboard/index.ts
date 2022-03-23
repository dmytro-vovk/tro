import App from '../../system';
import {$$, $html, $onClick, $text} from '../index';

type response = {
	message: string
}

type stream = {
	value: string
}

// Workaround for TS libs not knowing about crypto.randomUUID()
declare global {
	interface Crypto {
		randomUUID: () => string;
	}
}

export default class Dashboard {
	private qrData: HTMLInputElement;
	private qrImage: HTMLElement;

	constructor(private readonly app: App) {
	}

	public render(selector: string): void {
		$html(selector, `
			<div class="row">
				<div class="col-4">
					<div class="card card-primary card-outline card-outline-tabs">
						<div class="card-header">
							<h3 class="card-title">Приклад запиту/відповіді</h3>
						</div>
						<div class="card-body" id="response"></div>
						<div class="card-footer">
							<button id="ping" type="button" class="btn btn-success float-right">
								Натисни мене
							</button>
						</div>
					</div>
				</div>
				<div class="col-4">
					<div class="card card-primary card-outline card-outline-tabs">
						<div class="card-header">
							<h3 class="card-title">Серверний потік</h3>
						</div>
						<div class="card-body" id="stream"></div>
						<div class="card-footer"></div>
					</div>
				</div>
				<div class="col-4">
					<div class="card card-primary card-outline card-outline-tabs">
						<div class="card-header">
							<h3 class="card-title">QR Генератор</h3>
						</div>
						<div class="card-body" id="stream">
							<div class="input-group">
								<div class="input-group-prepend">
									<button id="gen-uuid" type="button" class="btn btn-default">UUID <i class="fa fa-arrow-right"></i></button>
								</div>
								<input type="text" class="form-control" id="data">
							</div>
							<div id="qr-image"></div>
						</div>
						<div class="card-footer">
							<button id="generate-qr" type="button" class="btn btn-success float-right">
								Згенерувати
							</button>
						</div>
					</div>
				</div>
			</div>`);

		this.qrData = $$('#data') as HTMLInputElement;
		this.qrImage = $$('#qr-image') as HTMLElement;

		$onClick('#gen-uuid', () => {
			this.qrData.value = crypto.randomUUID();
		});

		this.setupExampleHandler();
		this.subscribeToServerStream();
		this.setupQRGenerateHandler();
	}

	private subscribeToServerStream() {
		this.app.subscribe("example.stream", (data) => {
			const s = data as unknown as stream;
			$text("#stream", s.value);
		})
	}

	private setupQRGenerateHandler() {
		$onClick(
			"#generate-qr",
			() => {
				this.app.call(
						"code.generate_image",
						{
							data: this.qrData.value
						}
					)
					.then(
						(data) => {
							const qr = new Image();
							qr.src = 'data:image/png;base64,'+data;
							$text(this.qrImage, '');
							this.qrImage.appendChild(qr);
						},
						(error) => this.app.error(error),
					)
			}
		);
	}

	private setupExampleHandler() {
		$onClick(
			"#ping",
			() => {
				this.app.call(
						"example.method",
						{
							message: "привіт"
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
