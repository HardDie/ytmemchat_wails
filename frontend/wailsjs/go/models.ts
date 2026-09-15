export namespace main {
	
	export class SettingsForm {
	    streamId: string;
	    apiKey: string;
	    port: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.streamId = source["streamId"];
	        this.apiKey = source["apiKey"];
	        this.port = source["port"];
	    }
	}

}

