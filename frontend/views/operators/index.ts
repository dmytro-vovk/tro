import App from '../../system';
import {$html} from '../index';

export default class Operators {
	constructor(private readonly app: App) {
	}

	public render(selector: string): void {
		$html(selector, `
			<div class="row">
				<div class="col-12">
					<div class="card">
						<div class="card-header">
							<h3 class="card-title">Зареєстровані оператори</h3>
							<div class="card-tools">
								<div class="input-group">
									<button class="button button-xs">Додати</button>
								</div>
							</div>
						</div>
						<div class="card-body table-responsive p-0">
							<table class="table table-hover text-nowrap">
								<thead>
									<tr>
										<th>ID</th>
										<th>Логін</th>
									</tr>
								</thead>
								<tbody></tbody>
							</table>
						</div>
					</div>
				</div>
			</div>
		`)
	}
}
