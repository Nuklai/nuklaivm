export namespace backend {
	
	
	
	
	
	export class Config {
	    nuklaiRPC: string;
	    faucetRPC: string;
	    searchCores: number;
	    feedRPC: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nuklaiRPC = source["nuklaiRPC"];
	        this.faucetRPC = source["faucetRPC"];
	        this.searchCores = source["searchCores"];
	        this.feedRPC = source["feedRPC"];
	    }
	}
	
	
	
	
	

}

