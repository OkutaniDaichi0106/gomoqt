export interface ProfileInit {
    id: string;
}

export class Profile {
    readonly id: string;

    constructor(init: ProfileInit) {
        this.id = init.id;
    }

    
}