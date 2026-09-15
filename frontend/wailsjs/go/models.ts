export namespace main {
	
	export class OBSStatus {
	    listening: boolean;
	    error: string;
	    chatUrl: string;
	    overlayUrl: string;
	    indexUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new OBSStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.listening = source["listening"];
	        this.error = source["error"];
	        this.chatUrl = source["chatUrl"];
	        this.overlayUrl = source["overlayUrl"];
	        this.indexUrl = source["indexUrl"];
	    }
	}
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

