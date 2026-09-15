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
	export class RunStatus {
	    running: boolean;
	    connecting: boolean;
	    usingApiKey: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RunStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.connecting = source["connecting"];
	        this.usingApiKey = source["usingApiKey"];
	        this.error = source["error"];
	    }
	}
	export class SettingsForm {
	    streamId: string;
	    apiKey: string;
	    port: string;
	    ttsEnabled: boolean;
	    ttsVoiceName: string;
	    alertsEnabled: boolean;
	    alertsToken: string;
	    alertsMediaPath: string;
	    alertsCommandsFilePath: string;
	    webhookEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingsForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.streamId = source["streamId"];
	        this.apiKey = source["apiKey"];
	        this.port = source["port"];
	        this.ttsEnabled = source["ttsEnabled"];
	        this.ttsVoiceName = source["ttsVoiceName"];
	        this.alertsEnabled = source["alertsEnabled"];
	        this.alertsToken = source["alertsToken"];
	        this.alertsMediaPath = source["alertsMediaPath"];
	        this.alertsCommandsFilePath = source["alertsCommandsFilePath"];
	        this.webhookEnabled = source["webhookEnabled"];
	    }
	}
	export class StreamLookup {
	    streamId: string;
	    channelId: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new StreamLookup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.streamId = source["streamId"];
	        this.channelId = source["channelId"];
	        this.kind = source["kind"];
	    }
	}
	export class TTSVoice {
	    name: string;
	    languages: string;
	    gender: string;
	    details: string;
	
	    static createFrom(source: any = {}) {
	        return new TTSVoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.languages = source["languages"];
	        this.gender = source["gender"];
	        this.details = source["details"];
	    }
	}

}

