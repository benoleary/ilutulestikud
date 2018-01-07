import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { environment } from '../environments/environment';

@Injectable()
export class IlutulestikudService {
  httpClient: HttpClient;
  uriRoot: string;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
    this.uriRoot = environment.restRoot;
  }

  registeredPlayerNames(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "lobby/registered-player-names")
  }

  newPlayer(newPlayerName: string): Observable<any> {
    return this.httpClient.post(this.uriRoot + "lobby/new-player", {"Name": newPlayerName})
  }
}
