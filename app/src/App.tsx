import React from 'react';
import { SimulatorInstaller, Integration } from '@pinpt/agent.websdk';
import IntegrationUI from './integration';

function App() {
	// check to see if we are running local and need to run in simulation mode
	if (window === window.parent && window.location.href.indexOf('localhost') > 0) {
		const integration: Integration = {
			name: 'Pinpoint Software, Inc.',
			description: 'This is the Codefresh integration for Pinpoint',
			tags: [
				'CI/CD',
			],
			installed: false,
			refType: 'codefresh',
			icon: 'https://pinpoint.com/images/integrations/Codefresh.svg',
			publisher: {
				name: 'Pinpoint Software, Inc.',
				avatar: 'https://pinpoint.com/logo/logomark/blue.png',
				url: 'https://pinpoint.com'
			},
			uiURL: window.location.href
		};
		return <SimulatorInstaller integration={integration} />;
	}
	return (
		<div className="App">
			<IntegrationUI />
		</div>
	);
}

export default App;
