import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/Observable';
import { environment } from '../environments/environment';
import { BackendIdentification } from './models/backendidentification.model';
import { Player } from './models/player.model';

@Injectable()
export class IlutulestikudService {
  httpClient: HttpClient;
  uriRoot: string;

  constructor(httpClient: HttpClient) {
    this.httpClient = httpClient;
    this.uriRoot = environment.restRoot;
  }

  registeredPlayers(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "player/registered-players")
  }

  availableColors(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "player/available-colors")
  }

  newPlayer(newPlayerName: string): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "player/new-player",
      {
        "Name": newPlayerName
      })
  }

  updatePlayer(playerOverride: Player): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "player/update-player",
      {
        "Name": playerOverride.ForBackend.NameForPost,
        "Color": playerOverride.Color
      })
  }

  availableRulesets(): Observable<any> {
    return this.httpClient.get(this.uriRoot + "game/available-rulesets")
  }

  gamesWithPlayer(playerIdentification: BackendIdentification): Observable<any> {
    return this.httpClient.get(
      this.uriRoot
        + "game/all-games-with-player/"
        + encodeURIComponent(playerIdentification.IdentifierForGet))
  }

  newGame(
    newGameName: string,
    rulesetIdentifier: number,
    playerIdentifications: BackendIdentification[]): Observable<any> {
    const playerNames: string[] = playerIdentifications.map(playerIdentification => playerIdentification.NameForPost)
    return this.httpClient.post(
      this.uriRoot + "game/create-new-game",
      {
        "GameName": newGameName,
        "RulesetIdentifier": rulesetIdentifier,
        "PlayerNames": playerNames
      })
  }

  gameAsSeenByPlayer(
    gameIdentification: BackendIdentification,
    playerIdentification: BackendIdentification): Observable<any> {
    return this.httpClient.get(
      this.uriRoot 
        + "game/game-as-seen-by-player/"
        + encodeURIComponent(gameIdentification.IdentifierForGet)
        + "/" + encodeURIComponent(playerIdentification.IdentifierForGet))
  }

  sendChatMessage(
    gameIdentification: BackendIdentification,
    playerIdentification: BackendIdentification,
     chatMessage: string): Observable<any> {
    return this.httpClient.post(
      this.uriRoot + "game/record-chat-message",
      {
        "GameName": gameIdentification.NameForPost,
        "PlayerName": playerIdentification.NameForPost,
        "ChatMessage": chatMessage
      })
  }
}
