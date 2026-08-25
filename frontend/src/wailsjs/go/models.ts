export namespace sniffer {
	
	export class PcapDeviceInfo {
	    name: string;
	    pcapName: string;
	    description: string;
	    ips: string[];
	
	    static createFrom(source: any = {}) {
	        return new PcapDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.pcapName = source["pcapName"];
	        this.description = source["description"];
	        this.ips = source["ips"];
	    }
	}

}

export namespace usbserial {
	
	export class SerialDeviceInfo {
	    port_name: string;
	    is_usb: boolean;
	    serial_number: string;
	    product: string;
	    vendor_id: string;
	    product_id: string;
	    is_usb_device: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SerialDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port_name = source["port_name"];
	        this.is_usb = source["is_usb"];
	        this.serial_number = source["serial_number"];
	        this.product = source["product"];
	        this.vendor_id = source["vendor_id"];
	        this.product_id = source["product_id"];
	        this.is_usb_device = source["is_usb_device"];
	    }
	}

}

