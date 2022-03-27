import App from './system';
import Dashboard from './views/dashboard';
import Navigo from 'navigo';
import Page from './views/page';
import RPC from './system/rpc';
import Operators from './views/operators';

const contentSelector = '#content-wrapper';
const wsURL = window.location.protocol.replace(/^http/, 'ws') + '//' + window.location.host + '/ws';
const app = new App(new RPC(wsURL));
const homePage = new Page('Загальна Панель', new Dashboard(app));
const operatorsPage = new Page('Оператори', new Operators(app));

app.setRouter(new Navigo('/')
	.on('/', () => {
		app.sideBarToggle('/');
		homePage.render(contentSelector);
	})
	.on('/operators', () => {
		app.sideBarToggle('/operators');
		operatorsPage.render(contentSelector);
	})
);
