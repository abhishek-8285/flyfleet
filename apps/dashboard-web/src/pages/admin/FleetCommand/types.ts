export interface Vehicle {
    id: string;
    org_id: string;
    name: string;
    license_plate: string;
    make: string;
    model: string;
    year: number;
    fuel_type: string;
    vehicle_no: string;
    vin: string;
    odometer: number;
    service_due: string; // ISO date
    notes: string;
    status: 'active' | 'maintenance' | 'inactive';
    current_driver_id?: string;
    created_at: string;
    updated_at: string;
}

export interface Driver {
    id: string;
    first_name: string;
    last_name: string;
    full_name: string;
    status: string;
}

export interface Document {
    id: string;
    vehicle_id: string;
    type: string;
    doc_number: string;
    issue_date: string;
    expiry_date: string;
    file_url?: string;
    status: string;
    metadata?: any;
    created_at: string;
    updated_at: string;
}
