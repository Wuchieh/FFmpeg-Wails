export namespace backend {
	
	export class Task {
	    id: string;
	    type: string;
	    command: string;
	    status: string;
	    progress: number;
	    input: string;
	    output: string;
	    // Go type: time
	    createdAt: any;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.command = source["command"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.input = source["input"];
	        this.output = source["output"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.error = source["error"];
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

}

