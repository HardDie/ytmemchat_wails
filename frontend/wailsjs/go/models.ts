export namespace main {
	
	export class AlertCommand {
	    name: string;
	    file: string;
	    volume?: number;
	    scale?: number;
	
	    static createFrom(source: any = {}) {
	        return new AlertCommand(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.file = source["file"];
	        this.volume = source["volume"];
	        this.scale = source["scale"];
	    }
	}
	export class AlertCommandsFile {
	    path: string;
	    commands: AlertCommand[];
	
	    static createFrom(source: any = {}) {
	        return new AlertCommandsFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.commands = this.convertValues(source["commands"], AlertCommand);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
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
	    quotaUnits: number;
	    quotaUnitsLimit: number;
	    quotaSearch: number;
	    quotaSearchLimit: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RunStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.connecting = source["connecting"];
	        this.usingApiKey = source["usingApiKey"];
	        this.quotaUnits = source["quotaUnits"];
	        this.quotaUnitsLimit = source["quotaUnitsLimit"];
	        this.quotaSearch = source["quotaSearch"];
	        this.quotaSearchLimit = source["quotaSearchLimit"];
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
	    interruptHotkeyEnabled: boolean;
	    interruptHotkeyChord: string;
	    interruptHotkeyError: string;
	
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
	        this.interruptHotkeyEnabled = source["interruptHotkeyEnabled"];
	        this.interruptHotkeyChord = source["interruptHotkeyChord"];
	        this.interruptHotkeyError = source["interruptHotkeyError"];
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

