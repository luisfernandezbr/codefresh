import React, { useCallback, useState } from 'react';
import Loader from '@pinpt/uic.next/Loader';
import Icon from '@pinpt/uic.next/Icon';
import { faExclamationCircle } from '@fortawesome/free-solid-svg-icons';
import { useIntegration, Form, FormType, IAuth, IAPIKeyAuth, NoAction } from '@pinpt/agent.websdk';
import styles from './styles.module.less';

const Integration = () => {
	const { loading, config, setInstallEnabled, installed, setValidate } = useIntegration();
	const [validated, setValidated] = useState(false);
	const [error, setError] = useState('');
	const callback = useCallback(async (auth: IAuth | string) => {
		try {
			await setValidate({ ...config, apikey_auth: auth as IAPIKeyAuth });
			setInstallEnabled(true);
			setValidated(true);
			setError('');
		} catch (ex) {
			console.error(ex);
			setError(ex.message);
			setInstallEnabled(false);
			setValidated(false);
		}
	}, []);
	if (loading) {
		return <Loader centered />;
	}
	if (!config.apikey_auth?.apikey || !installed) {
		return (
			<div className={styles.Container}>
				<Form
					type={FormType.API}
					name="Codefresh"
					form={{ url: { disabled: true } }}
					callback={callback}
					title="Connect Codefresh to Pinpoint."
					enabledValidator={async (auth: IAuth | string) => {
						const _auth = auth as any;
						return _auth?.apikey?.length >= 48;
					}}
					button={validated ? 'Revalidate' : 'Validate'}
					intro={
						<>
							<ul>
								<li>Please provide the API Key to connect to your Codefresh instance. You can create an API Key by visiting the <a target="_blank" rel="noopener" href="https://g.codefresh.io/user/settings">User Settings</a> in Codefresh.</li>
								<li>Your API Key only needs the read-only build scope.</li>
							</ul>
							{error && (
								<div className={styles.Error}>
									<Icon icon={faExclamationCircle} /> {error}
								</div>
							)}
						</>
					}
				/>
			</div>
		);
	}
	return (
		<div className={styles.Container}>
			<NoAction />
		</div>
	);
};

export default Integration;
